package main

import (
	context "context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"gitlab.vk-golang.com/vk-golang/lectures/08_microservices/99_hw/microservice/domain"
	"gitlab.vk-golang.com/vk-golang/lectures/08_microservices/99_hw/microservice/services/acl"
	"gitlab.vk-golang.com/vk-golang/lectures/08_microservices/99_hw/microservice/services/logger"
	"google.golang.org/grpc"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные
// если хочется, то для красоты можно разнести логику по разным файликам

type AdminService struct {
	UnimplementedAdminServer
	ctx         context.Context
	logRegistry *logger.LogChannelRegistry
}

func (s *AdminService) Logging(n *Nothing, stream Admin_LoggingServer) error {
	localChan := make(chan *domain.Event, 100)
	s.logRegistry.AddChannel(localChan)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case evt := <-localChan:
				fmt.Println("got some", evt)
				stream.Send(&Event{
					Timestamp: evt.Timestamp,
					Consumer:  evt.Consumer,
					Method:    evt.Method,
					Host:      evt.Host,
				})
			}
		}
	}()
	<-s.ctx.Done()
	s.logRegistry.RemoveChannel(localChan)
	close(localChan)
	return nil
}

func (s *AdminService) Statistics(interval *StatInterval, stream Admin_StatisticsServer) error {
	localChan := make(chan *domain.Event, 100)
	s.logRegistry.AddChannel(localChan)
	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds * uint64(time.Second)))
	events := make([]*domain.Event, 0)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("tick", events)
				stream.Send(&Stat{
					ByMethod:   statsByMethod(events),
					ByConsumer: statsByConsumer(events),
					Timestamp:  time.Now().Unix(),
				})
				events = []*domain.Event{}
			case <-s.ctx.Done():
				return
			case ev := <-localChan:
				events = append(events, ev)
			}
		}
	}()
	<-s.ctx.Done()
	s.logRegistry.RemoveChannel(localChan)
	close(localChan)
	return nil
}

func statsByMethod(events []*domain.Event) map[string]uint64 {
	byMethod := make(map[string]uint64)
	for _, evt := range events {
		byMethod[evt.Method] += 1
	}
	fmt.Println("BY METHOD", byMethod)
	return byMethod
}

func statsByConsumer(events []*domain.Event) map[string]uint64 {
	ByConsumer := make(map[string]uint64)
	for _, evt := range events {
		ByConsumer[evt.Consumer] += 1
	}
	return ByConsumer
}

type DomainService struct {
	UnimplementedBizServer
}

func (d *DomainService) Check(context.Context, *Nothing) (*Nothing, error) { return &Nothing{}, nil }
func (d *DomainService) Add(context.Context, *Nothing) (*Nothing, error)   { return &Nothing{}, nil }
func (d *DomainService) Test(context.Context, *Nothing) (*Nothing, error)  { return &Nothing{}, nil }

func StartMyMicroservice(ctx context.Context, addr string, incAcl string) error {
	aclData, err := parseAcl(incAcl)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logRegistry := logger.NewChannelRegistry()
	logInterceptor := logger.NewLoggerInterceptor(logRegistry)
	aclService := acl.NewAclService(aclData)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			aclService.AclInterceptor,
			logInterceptor.New,
		),
		grpc.ChainStreamInterceptor(
			aclService.AclStreamInterceptor,
			logInterceptor.Stream,
		),
	)

	admin := AdminService{
		ctx:         ctx,
		logRegistry: logRegistry,
	}

	RegisterAdminServer(server, &admin)
	RegisterBizServer(server, &DomainService{})

	go server.Serve(listener)
	go func() {
		<-ctx.Done()
		fmt.Println("graceful shutdown")
		server.GracefulStop()
		listener.Close()
	}()

	return nil
}

func parseAcl(incAcl string) (acl.AclData, error) {
	aclData := &acl.AclData{}
	err := json.Unmarshal([]byte(incAcl), aclData)
	if err != nil {
		return acl.AclData{}, err
	}

	return *aclData, nil
}
