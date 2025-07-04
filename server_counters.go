package msmtpd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

type wrapper struct {
	Original io.ReadWriteCloser
	server   *Server
}

func (w *wrapper) Read(p []byte) (n int, err error) {
	n, err = w.Original.Read(p)
	atomic.AddUint64(&w.server.bytesRead, uint64(n))
	return
}

func (w *wrapper) Write(p []byte) (n int, err error) {
	n, err = w.Original.Write(p)
	atomic.AddUint64(&w.server.bytesWritten, uint64(n))
	return
}

func (w *wrapper) Close() error {
	return w.Original.Close()
}

func (srv *Server) wrapWithCounters(stream io.ReadWriteCloser) (wrapped io.ReadWriteCloser) {
	return &wrapper{Original: stream, server: srv}
}

// GetBytesWritten returns number of bytes written
func (srv *Server) GetBytesWritten() uint64 {
	return srv.bytesWritten
}

// GetBytesRead returns number of bytes written
func (srv *Server) GetBytesRead() uint64 {
	return srv.bytesRead
}

// GetTransactionsCount returns number of all transactions this server processed
func (srv *Server) GetTransactionsCount() uint64 {
	return srv.transactionsAll
}

// GetActiveTransactionsCount returns number of active transactions this server is processing
func (srv *Server) GetActiveTransactionsCount() int32 {
	return srv.transactionsActive
}

// GetSuccessfulTransactionsCount returns number of successful transactions this server processed
func (srv *Server) GetSuccessfulTransactionsCount() uint64 {
	return srv.transactionsSuccess
}

// GetFailedTransactionsCount returns number of failed transactions this server processed
func (srv *Server) GetFailedTransactionsCount() uint64 {
	return srv.transactionsFail
}

// ResetCounters resets counters
func (srv *Server) ResetCounters() {
	srv.bytesRead = 0
	srv.bytesWritten = 0
	srv.transactionsAll = 0
}

// StartPrometheusScrapperEndpoint starts prometheus scrapper endpoint with data
// in this format https://prometheus.io/docs/instrumenting/exposition_formats/
func (srv *Server) StartPrometheusScrapperEndpoint(address, path string) (err error) {
	httpServ := http.Server{
		Addr:    address,
		Handler: http.DefaultServeMux,
	}
	http.HandleFunc(path, func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "text/plain; version=0.0.4")
		res.WriteHeader(http.StatusOK)
		fmt.Fprintf(res, "bytes_read{hostname=\"%s\"} %v %v\n",
			srv.Hostname, srv.GetBytesRead(), srv.lastTransactionStartedAt.UnixMilli())
		fmt.Fprintf(res, "bytes_written{hostname=\"%s\"} %v %v\n",
			srv.Hostname, srv.GetBytesWritten(), srv.lastTransactionStartedAt.UnixMilli())
		fmt.Fprintf(res, "active_transactions_count{hostname=\"%s\"} %v %v\n",
			srv.Hostname, srv.GetActiveTransactionsCount(), srv.lastTransactionStartedAt.UnixMilli())
		fmt.Fprintf(res, "all_transactions_count{hostname=\"%s\"} %v %v\n",
			srv.Hostname, srv.GetTransactionsCount(), srv.lastTransactionStartedAt.UnixMilli())
		fmt.Fprintf(res, "successfull_transactions_count{hostname=\"%s\"} %v %v\n",
			srv.Hostname, srv.GetSuccessfulTransactionsCount(), srv.lastTransactionStartedAt.UnixMilli())
		fmt.Fprintf(res, "failed_transactions_count{hostname=\"%s\"} %v %v\n",
			srv.Hostname, srv.GetFailedTransactionsCount(), srv.lastTransactionStartedAt.UnixMilli())
	})
	go func() {
		<-srv.Context.Done()
		httpServ.Close()
	}()
	err = httpServ.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
	}
	return err
}
