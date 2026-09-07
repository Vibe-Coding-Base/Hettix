package api

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

func HTTPHandler(resolver *Resolver, gqlEndpoint string) http.Handler {
	gqlServer := handler.NewDefaultServer(NewExecutableSchema(Config{
		Resolvers: resolver,
	}))
	pgHandler := playground.Handler("GraphQL Playground", gqlEndpoint)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			gqlServer.ServeHTTP(w, r)
		case http.MethodGet:
			pgHandler.ServeHTTP(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
