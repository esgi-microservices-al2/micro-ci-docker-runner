package main

import (
	"fmt"
	"log"
	"time"

	consul "github.com/hashicorp/consul/api"
)

type Service struct {
	ID          string
	Name        string
	TTL         time.Duration
	consul      *consul.Client
	ConsulAgent *consul.Agent
}

// NewConsul returns a Client interface for given consul address
func NewConsulClient(addr string, token string) (*Service, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	config.Token = token
	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Service{
		ID:     "docker-runner",
		Name:   "docker-runner",
		TTL:    time.Second * 15,
		consul: c,
	}, nil
}

// Register a service with consul local agent
func (s *Service) Register() error {
	serviceDef := &consul.AgentServiceRegistration{
		ID:   s.Name,
		Name: s.Name,
		Check: &consul.AgentServiceCheck{
			TTL: fmt.Sprintf("%fs", s.TTL.Seconds()),
		},
	}
	s.ConsulAgent = s.consul.Agent()
	if err := s.ConsulAgent.ServiceRegister(serviceDef); err != nil {
		return err
	}
	go s.UpdateTTL()
	return nil
}

// update TTL with clock
func (s *Service) UpdateTTL() {
	ticker := time.NewTicker(s.TTL / 2)
	for range ticker.C {
		s.update()
	}
}

// update TTL
func (s *Service) update() {
	if agentErr := s.ConsulAgent.UpdateTTL("service:"+s.ID, "", "pass"); agentErr != nil {
		log.Print(agentErr)
	}
}

// DeRegister a service with consul local agent
func (s *Service) DeRegister(id string) error {
	return s.consul.Agent().ServiceDeregister(id)
}

// Service return a service
func (s *Service) Service(service, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := s.consul.Health().Service(service, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", service)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}
