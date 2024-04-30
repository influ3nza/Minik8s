package application

import "net/http"

type Handler func(response http.ResponseWriter, req *http.Request)

func Run(httpFunc map[string]Handler) {
	for method, handler := range httpFunc {
		http.HandleFunc(method, handler)
	}

	err := http.ListenAndServe(":42100", nil)
	if err != nil {
		return
	}
}
