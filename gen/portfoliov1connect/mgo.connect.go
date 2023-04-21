// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: mgo.proto

package portfoliov1connect

import (
	context "context"
	errors "errors"
	connect_go "github.com/bufbuild/connect-go"
	gen "github.com/oxisto/money-gopher/gen"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect_go.IsAtLeastVersion1_7_0

const (
	// PortfolioServiceName is the fully-qualified name of the PortfolioService service.
	PortfolioServiceName = "mgo.portfolio.v1.PortfolioService"
	// SecuritiesServiceName is the fully-qualified name of the SecuritiesService service.
	SecuritiesServiceName = "mgo.portfolio.v1.SecuritiesService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// PortfolioServiceCreatePortfolioProcedure is the fully-qualified name of the PortfolioService's
	// CreatePortfolio RPC.
	PortfolioServiceCreatePortfolioProcedure = "/mgo.portfolio.v1.PortfolioService/CreatePortfolio"
	// SecuritiesServiceListSecuritiesProcedure is the fully-qualified name of the SecuritiesService's
	// ListSecurities RPC.
	SecuritiesServiceListSecuritiesProcedure = "/mgo.portfolio.v1.SecuritiesService/ListSecurities"
	// SecuritiesServiceGetSecurityProcedure is the fully-qualified name of the SecuritiesService's
	// GetSecurity RPC.
	SecuritiesServiceGetSecurityProcedure = "/mgo.portfolio.v1.SecuritiesService/GetSecurity"
	// SecuritiesServiceCreateSecurityProcedure is the fully-qualified name of the SecuritiesService's
	// CreateSecurity RPC.
	SecuritiesServiceCreateSecurityProcedure = "/mgo.portfolio.v1.SecuritiesService/CreateSecurity"
	// SecuritiesServiceUpdateSecurityProcedure is the fully-qualified name of the SecuritiesService's
	// UpdateSecurity RPC.
	SecuritiesServiceUpdateSecurityProcedure = "/mgo.portfolio.v1.SecuritiesService/UpdateSecurity"
	// SecuritiesServiceDeleteSecurityProcedure is the fully-qualified name of the SecuritiesService's
	// DeleteSecurity RPC.
	SecuritiesServiceDeleteSecurityProcedure = "/mgo.portfolio.v1.SecuritiesService/DeleteSecurity"
)

// PortfolioServiceClient is a client for the mgo.portfolio.v1.PortfolioService service.
type PortfolioServiceClient interface {
	CreatePortfolio(context.Context, *connect_go.Request[gen.PortfolioCreateMessage]) (*connect_go.Response[gen.Portfolio], error)
}

// NewPortfolioServiceClient constructs a client for the mgo.portfolio.v1.PortfolioService service.
// By default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped
// responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewPortfolioServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) PortfolioServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &portfolioServiceClient{
		createPortfolio: connect_go.NewClient[gen.PortfolioCreateMessage, gen.Portfolio](
			httpClient,
			baseURL+PortfolioServiceCreatePortfolioProcedure,
			opts...,
		),
	}
}

// portfolioServiceClient implements PortfolioServiceClient.
type portfolioServiceClient struct {
	createPortfolio *connect_go.Client[gen.PortfolioCreateMessage, gen.Portfolio]
}

// CreatePortfolio calls mgo.portfolio.v1.PortfolioService.CreatePortfolio.
func (c *portfolioServiceClient) CreatePortfolio(ctx context.Context, req *connect_go.Request[gen.PortfolioCreateMessage]) (*connect_go.Response[gen.Portfolio], error) {
	return c.createPortfolio.CallUnary(ctx, req)
}

// PortfolioServiceHandler is an implementation of the mgo.portfolio.v1.PortfolioService service.
type PortfolioServiceHandler interface {
	CreatePortfolio(context.Context, *connect_go.Request[gen.PortfolioCreateMessage]) (*connect_go.Response[gen.Portfolio], error)
}

// NewPortfolioServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewPortfolioServiceHandler(svc PortfolioServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle(PortfolioServiceCreatePortfolioProcedure, connect_go.NewUnaryHandler(
		PortfolioServiceCreatePortfolioProcedure,
		svc.CreatePortfolio,
		opts...,
	))
	return "/mgo.portfolio.v1.PortfolioService/", mux
}

// UnimplementedPortfolioServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedPortfolioServiceHandler struct{}

func (UnimplementedPortfolioServiceHandler) CreatePortfolio(context.Context, *connect_go.Request[gen.PortfolioCreateMessage]) (*connect_go.Response[gen.Portfolio], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("mgo.portfolio.v1.PortfolioService.CreatePortfolio is not implemented"))
}

// SecuritiesServiceClient is a client for the mgo.portfolio.v1.SecuritiesService service.
type SecuritiesServiceClient interface {
	ListSecurities(context.Context, *connect_go.Request[gen.ListSecuritiesRequest]) (*connect_go.Response[gen.ListSecuritiesResponse], error)
	GetSecurity(context.Context, *connect_go.Request[gen.GetSecurityRequest]) (*connect_go.Response[gen.Security], error)
	CreateSecurity(context.Context, *connect_go.Request[gen.CreateSecurityRequest]) (*connect_go.Response[gen.Security], error)
	UpdateSecurity(context.Context, *connect_go.Request[gen.UpdateSecurityRequest]) (*connect_go.Response[gen.Security], error)
	DeleteSecurity(context.Context, *connect_go.Request[gen.DeleteSecurityRequest]) (*connect_go.Response[emptypb.Empty], error)
}

// NewSecuritiesServiceClient constructs a client for the mgo.portfolio.v1.SecuritiesService
// service. By default, it uses the Connect protocol with the binary Protobuf Codec, asks for
// gzipped responses, and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply
// the connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewSecuritiesServiceClient(httpClient connect_go.HTTPClient, baseURL string, opts ...connect_go.ClientOption) SecuritiesServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &securitiesServiceClient{
		listSecurities: connect_go.NewClient[gen.ListSecuritiesRequest, gen.ListSecuritiesResponse](
			httpClient,
			baseURL+SecuritiesServiceListSecuritiesProcedure,
			connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
			connect_go.WithClientOptions(opts...),
		),
		getSecurity: connect_go.NewClient[gen.GetSecurityRequest, gen.Security](
			httpClient,
			baseURL+SecuritiesServiceGetSecurityProcedure,
			connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
			connect_go.WithClientOptions(opts...),
		),
		createSecurity: connect_go.NewClient[gen.CreateSecurityRequest, gen.Security](
			httpClient,
			baseURL+SecuritiesServiceCreateSecurityProcedure,
			opts...,
		),
		updateSecurity: connect_go.NewClient[gen.UpdateSecurityRequest, gen.Security](
			httpClient,
			baseURL+SecuritiesServiceUpdateSecurityProcedure,
			opts...,
		),
		deleteSecurity: connect_go.NewClient[gen.DeleteSecurityRequest, emptypb.Empty](
			httpClient,
			baseURL+SecuritiesServiceDeleteSecurityProcedure,
			opts...,
		),
	}
}

// securitiesServiceClient implements SecuritiesServiceClient.
type securitiesServiceClient struct {
	listSecurities *connect_go.Client[gen.ListSecuritiesRequest, gen.ListSecuritiesResponse]
	getSecurity    *connect_go.Client[gen.GetSecurityRequest, gen.Security]
	createSecurity *connect_go.Client[gen.CreateSecurityRequest, gen.Security]
	updateSecurity *connect_go.Client[gen.UpdateSecurityRequest, gen.Security]
	deleteSecurity *connect_go.Client[gen.DeleteSecurityRequest, emptypb.Empty]
}

// ListSecurities calls mgo.portfolio.v1.SecuritiesService.ListSecurities.
func (c *securitiesServiceClient) ListSecurities(ctx context.Context, req *connect_go.Request[gen.ListSecuritiesRequest]) (*connect_go.Response[gen.ListSecuritiesResponse], error) {
	return c.listSecurities.CallUnary(ctx, req)
}

// GetSecurity calls mgo.portfolio.v1.SecuritiesService.GetSecurity.
func (c *securitiesServiceClient) GetSecurity(ctx context.Context, req *connect_go.Request[gen.GetSecurityRequest]) (*connect_go.Response[gen.Security], error) {
	return c.getSecurity.CallUnary(ctx, req)
}

// CreateSecurity calls mgo.portfolio.v1.SecuritiesService.CreateSecurity.
func (c *securitiesServiceClient) CreateSecurity(ctx context.Context, req *connect_go.Request[gen.CreateSecurityRequest]) (*connect_go.Response[gen.Security], error) {
	return c.createSecurity.CallUnary(ctx, req)
}

// UpdateSecurity calls mgo.portfolio.v1.SecuritiesService.UpdateSecurity.
func (c *securitiesServiceClient) UpdateSecurity(ctx context.Context, req *connect_go.Request[gen.UpdateSecurityRequest]) (*connect_go.Response[gen.Security], error) {
	return c.updateSecurity.CallUnary(ctx, req)
}

// DeleteSecurity calls mgo.portfolio.v1.SecuritiesService.DeleteSecurity.
func (c *securitiesServiceClient) DeleteSecurity(ctx context.Context, req *connect_go.Request[gen.DeleteSecurityRequest]) (*connect_go.Response[emptypb.Empty], error) {
	return c.deleteSecurity.CallUnary(ctx, req)
}

// SecuritiesServiceHandler is an implementation of the mgo.portfolio.v1.SecuritiesService service.
type SecuritiesServiceHandler interface {
	ListSecurities(context.Context, *connect_go.Request[gen.ListSecuritiesRequest]) (*connect_go.Response[gen.ListSecuritiesResponse], error)
	GetSecurity(context.Context, *connect_go.Request[gen.GetSecurityRequest]) (*connect_go.Response[gen.Security], error)
	CreateSecurity(context.Context, *connect_go.Request[gen.CreateSecurityRequest]) (*connect_go.Response[gen.Security], error)
	UpdateSecurity(context.Context, *connect_go.Request[gen.UpdateSecurityRequest]) (*connect_go.Response[gen.Security], error)
	DeleteSecurity(context.Context, *connect_go.Request[gen.DeleteSecurityRequest]) (*connect_go.Response[emptypb.Empty], error)
}

// NewSecuritiesServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewSecuritiesServiceHandler(svc SecuritiesServiceHandler, opts ...connect_go.HandlerOption) (string, http.Handler) {
	mux := http.NewServeMux()
	mux.Handle(SecuritiesServiceListSecuritiesProcedure, connect_go.NewUnaryHandler(
		SecuritiesServiceListSecuritiesProcedure,
		svc.ListSecurities,
		connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
		connect_go.WithHandlerOptions(opts...),
	))
	mux.Handle(SecuritiesServiceGetSecurityProcedure, connect_go.NewUnaryHandler(
		SecuritiesServiceGetSecurityProcedure,
		svc.GetSecurity,
		connect_go.WithIdempotency(connect_go.IdempotencyNoSideEffects),
		connect_go.WithHandlerOptions(opts...),
	))
	mux.Handle(SecuritiesServiceCreateSecurityProcedure, connect_go.NewUnaryHandler(
		SecuritiesServiceCreateSecurityProcedure,
		svc.CreateSecurity,
		opts...,
	))
	mux.Handle(SecuritiesServiceUpdateSecurityProcedure, connect_go.NewUnaryHandler(
		SecuritiesServiceUpdateSecurityProcedure,
		svc.UpdateSecurity,
		opts...,
	))
	mux.Handle(SecuritiesServiceDeleteSecurityProcedure, connect_go.NewUnaryHandler(
		SecuritiesServiceDeleteSecurityProcedure,
		svc.DeleteSecurity,
		opts...,
	))
	return "/mgo.portfolio.v1.SecuritiesService/", mux
}

// UnimplementedSecuritiesServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedSecuritiesServiceHandler struct{}

func (UnimplementedSecuritiesServiceHandler) ListSecurities(context.Context, *connect_go.Request[gen.ListSecuritiesRequest]) (*connect_go.Response[gen.ListSecuritiesResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("mgo.portfolio.v1.SecuritiesService.ListSecurities is not implemented"))
}

func (UnimplementedSecuritiesServiceHandler) GetSecurity(context.Context, *connect_go.Request[gen.GetSecurityRequest]) (*connect_go.Response[gen.Security], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("mgo.portfolio.v1.SecuritiesService.GetSecurity is not implemented"))
}

func (UnimplementedSecuritiesServiceHandler) CreateSecurity(context.Context, *connect_go.Request[gen.CreateSecurityRequest]) (*connect_go.Response[gen.Security], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("mgo.portfolio.v1.SecuritiesService.CreateSecurity is not implemented"))
}

func (UnimplementedSecuritiesServiceHandler) UpdateSecurity(context.Context, *connect_go.Request[gen.UpdateSecurityRequest]) (*connect_go.Response[gen.Security], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("mgo.portfolio.v1.SecuritiesService.UpdateSecurity is not implemented"))
}

func (UnimplementedSecuritiesServiceHandler) DeleteSecurity(context.Context, *connect_go.Request[gen.DeleteSecurityRequest]) (*connect_go.Response[emptypb.Empty], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("mgo.portfolio.v1.SecuritiesService.DeleteSecurity is not implemented"))
}
