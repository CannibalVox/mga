// +build !ignore_autogenerated

// Copyright 2020 Acme Inc.
// All rights reserved.
//
// Licensed under "Only for testing purposes" license.

// Code generated by mga tool. DO NOT EDIT.

package pkgdriver

import (
	"context"
	"errors"
	"github.com/go-kit/kit/endpoint"
	kitoc "github.com/go-kit/kit/tracing/opencensus"
	kitxendpoint "github.com/sagikazarmark/kitx/endpoint"
	"sagikazarmark.dev/mga/internal/generate/kit/endpoint/testdata/generator/service_with_struct"
)

// endpointError identifies an error that should be returned as an error endpoint.
type endpointError interface {
	EndpointError() bool
}

// Endpoints collects all of the endpoints that compose the underlying service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	CreateTodo endpoint.Endpoint
}

// MakeEndpoints returns a(n) Endpoints struct where each endpoint invokes
// the corresponding method on the provided service.
func MakeEndpoints(service service_with_struct.Service, middleware ...endpoint.Middleware) Endpoints {
	mw := kitxendpoint.Combine(middleware...)

	return Endpoints{CreateTodo: kitxendpoint.OperationNameMiddleware("service_with_struct.CreateTodo")(mw(MakeCreateTodoEndpoint(service)))}
}

// TraceEndpoints returns a(n) Endpoints struct where each endpoint is wrapped with a tracing middleware.
func TraceEndpoints(endpoints Endpoints) Endpoints {
	return Endpoints{CreateTodo: kitoc.TraceEndpoint("service_with_struct.CreateTodo")(endpoints.CreateTodo)}
}

// CreateTodoRequest is a request struct for CreateTodo endpoint.
type CreateTodoRequest struct {
	NewTodo service_with_struct.NewTodo
}

// CreateTodoResponse is a response struct for CreateTodo endpoint.
type CreateTodoResponse struct {
	Response service_with_struct.CreatedTodo
	Err      error
}

// MakeCreateTodoEndpoint returns an endpoint for the matching method of the underlying service.
func MakeCreateTodoEndpoint(service service_with_struct.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*CreateTodoRequest)

		response, err := service.CreateTodo(ctx, req.NewTodo)

		if err != nil {
			if endpointErr := endpointError(nil); errors.As(err, &endpointErr) && endpointErr.EndpointError() {
				return &CreateTodoResponse{
					Err:      err,
					Response: response,
				}, err
			}

			return &CreateTodoResponse{
				Err:      err,
				Response: response,
			}, nil
		}

		return &CreateTodoResponse{Response: response}, nil
	}
}
