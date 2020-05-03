package appliance

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gregorychen3/espresso-controller/internal/log"
	"github.com/gregorychen3/espresso-controller/internal/metrics"
	"github.com/gregorychen3/espresso-controller/pkg/appliancepb"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Configuration struct {
	Port        int
	RelayPinNum int
}

type Server struct {
	c Configuration

	grpcApplianceServer appliancepb.ApplianceServer
	grpcServer          *grpc.Server

	shutdownCh chan struct{}
}

func New(c Configuration) *Server {
	return &Server{
		c:          c,
		shutdownCh: make(chan struct{}),
	}
}

func (s *Server) Run() error {
	if err := metrics.InitMetrics(); err != nil {
		return errors.Wrap(err, "initializing metrics")
	}

	grpcController, err := newGrpcController(s.c)
	if err != nil {
		return err
	}
	s.grpcApplianceServer = grpcController

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(log.Logger),
		)))
	s.grpcServer = grpcServer

	go s.serveTCP()

	s.watchSignals() // blocks until signal received

	if err := s.Shutdown(); err != nil {
		log.Error("Failed while shutting down", zap.Error(err))
		return err
	} else {
		log.Info("Shutdown complete")
		return nil
	}
}

func (s *Server) serveTCP() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.c.Port))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to listen on port %d", s.c.Port))
	}

	appliancepb.RegisterApplianceServer(s.grpcServer, s.grpcApplianceServer)

	mux := cmux.New(listener)
	grpcListener := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	http1Listener := mux.Match(cmux.HTTP1())

	eg := errgroup.Group{}
	eg.Go(func() error { return s.serveGRPC(grpcListener, s.grpcServer) })
	eg.Go(func() error { return s.serveHTTP1(http1Listener, s.grpcServer) })
	eg.Go(func() error { return mux.Serve() })
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *Server) serveGRPC(listener net.Listener, grpcServer *grpc.Server) error {
	log.Info("Initializing gRPC server", zap.Int("port", s.c.Port))
	if err := grpcServer.Serve(listener); err != nil {
		return errors.Wrap(err, "gRPC server failed")
	}
	return nil
}

func (s *Server) serveHTTP1(listener net.Listener, grpcServer *grpc.Server) error {
	log.Info("Initializing gRPC web server", zap.Int("port", s.c.Port))
	server := NewGRPCWebServer(grpcServer)
	if err := server.Listen(listener, true /*TODO*/); err != nil {
		log.Error("gRPC web server failed", zap.Error(err))
		return errors.Wrap(err, "gRPC web server failed")
	}
	return nil
}

func (s *Server) watchSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP)

	sig := <-sigCh
	log.Info("Received signal", zap.Stringer("signal", sig))
}

func (s *Server) Shutdown() error {
	return nil
}
