package http

type Route struct {
	Method     Method
	Path       string
	Handler    HandlerFunc
	Middleware []Middleware
}

func (r Route) HandlerFunc(
	globalMiddleware ...Middleware,
) HandlerFunc {
	middleware := make(
		[]Middleware,
		0,
		len(globalMiddleware)+len(r.Middleware),
	)

	middleware = append(middleware, globalMiddleware...)
	middleware = append(middleware, r.Middleware...)

	return ChainMiddleware(r.Handler, middleware...)
}
