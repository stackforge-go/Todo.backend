package http

type APIVersion string

type Router struct {
	apiVersion APIVersion
	routes     []Route
	middleware []Middleware
}

func NewRouter(
	apiVersion APIVersion,
	middleware ...Middleware,
) *Router {
	return &Router{
		apiVersion: apiVersion,
		middleware: middleware,
	}
}

func (r *Router) AddRoute(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Router) AddRoutes(routes []Route) {
	for _, route := range routes {
		r.AddRoute(route)
	}
}
