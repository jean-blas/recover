package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(":3000", recoverDemo(mux)))
}

// Here we will replace the original mux with our, which adds a defer function that allows to recover from panic
// Indeed, our mux is an http.Handler (see ListenAndServe)
// hence we need a function that uses and returns an http.Handler
func recoverDemo(app http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add a defer function to the application without modifying the already existing HandleFunc
		defer func() {
			if res := recover(); res != nil {
				fmt.Println("Recover in panicAfterDemo")
				http.Error(w, "Something went wrong!", http.StatusInternalServerError)

				strace := string(debug.Stack())
				log.Println(strace) // log the stacktrace

				fmt.Fprintf(w, "Panic : %s<p><pre>%+v</pre>", res, strace) // Write the stacktrace in UI
			}
		}()

		// Replace the original ResponseWriter with our myResponseWriter
		mrw := &myResponseWriter{ResponseWriter: w}
		app.ServeHTTP(mrw, r) // Serve the application
		mrw.flush()
	}
}

// Here we will replace the original ResponseWriter with our
// which will overwrite some functions we want to achieve our goal
type myResponseWriter struct {
	// Header() Header
	// Write([]byte) (int, error)
	// WriteHeader(statusCode int)
	http.ResponseWriter
	writes     [][]byte
	statusCode int
}

func (m *myResponseWriter) Write(b []byte) (int, error) {
	m.writes = append(m.writes, b)
	return len(b), nil
}

func (m *myResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

// transfer data from myResponseWriter to the original ResponseWriter
func (m *myResponseWriter) flush() {
	if m.statusCode != 0 {
		m.ResponseWriter.WriteHeader(m.statusCode)
	}
	for _, b := range m.writes {
		m.ResponseWriter.Write(b)
	}
}

//********************* Original code *************************//
func panicDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

func funcThatPanics() {
	panic("Oh no!")
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}
