package rabbitmq

type Router struct {
	routes     []Route
	middleware []Middleware
}

func NewRouter(middleware ...Middleware) *Router {
	return &Router{middleware: middleware}
}

func (r *Router) AddRoute(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Router) AddRoutes(routes ...Route) {
	r.routes = append(r.routes, routes...)
}
