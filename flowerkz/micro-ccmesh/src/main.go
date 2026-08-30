package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"ccmeshclient/pkg/ccmesh"

	"cs.utexas.edu/zjia/faas"
	"cs.utexas.edu/zjia/faas/types"
)

type Handler struct {
	env types.Environment
}

type funcHandlerFactory struct {
}

func (f *funcHandlerFactory) New(env types.Environment, funcName string) (types.FuncHandler, error) {
	if funcName == "Entry1" {
		return &Handler{env: env}, nil
	} else if funcName == "Entry2" {
		return &Handler{env: env}, nil
	} else if funcName == "Entry3" {
		return &Handler{env: env}, nil
	} else if funcName == "Entry4" {
		return &Handler{env: env}, nil
	} else if funcName == "Entry5" {
		return &Handler{env: env}, nil
	} else if funcName == "Sink" {
		return &Handler{env: env}, nil
	} else {
		return nil, nil
	}
}

func (f *funcHandlerFactory) GrpcNew(env types.Environment, service string) (types.GrpcFuncHandler, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (h *Handler) Call(ctx context.Context, input []byte) ([]byte, error) {
	res := ccmesh.Run(input)
	return res, nil
}
func monitorGoroutines(pid string) error {
	file, err := os.Create(pid + "_routines.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "t,nb_goroutines")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	start := time.Now()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(start).Seconds()
			n := runtime.NumGoroutine()

			if _, err := fmt.Fprintf(file, "%.0f,%d\n", elapsed, n); err != nil {
				return err
			}
		}
	}
}

func main() {
	faas.Serve(&funcHandlerFactory{})

	pid := strconv.Itoa(os.Getpid())

	go func() {
		monitorGoroutines(pid)
	}()
}
