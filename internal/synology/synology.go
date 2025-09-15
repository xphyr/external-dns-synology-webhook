package synology

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"

	"github.com/xphyr/synology-csi/pkg/dsm/webapi"
)

type SynologyProvider struct {
	provider.BaseProvider
	client           *webapi.DSM
	zoneIDNameMapper provider.ZoneIDName
	domainFilter     endpoint.DomainFilter
	DryRun           bool
}

// Configuration contains the Synology provider's configuration.
type Configuration struct {
	SynologyHostName   string `env:"Synology_HOST_NAME" required:"true"`
	SynologyPortNumber string `env:"Synology_HOST_NAME" default:"5001"`
	SynologyUsername   string `env:"Synology_USERNAME" required:"true"`
	SynologyPassword   string `env:"Synology_PASSWORD" required:"true"`
	// If set to true, no changes will be applied to the DNS records.
	// Instead, the changes will be logged.
	DryRun     bool     `env:"DRY_RUN" default:"false"`
	DomainList []string `env:"DOMAIN_FILTER" default:""`
}

func NewProvider(providerConfig *Configuration) *SynologyProvider {

	client := &webapi.DSM{
		Ip:       providerConfig.SynologyHostName,
		Username: providerConfig.SynologyUsername,
		Password: providerConfig.SynologyPassword,
		Port:     5001,
		Https:    true,
	}

	err := client.Login()
	if err != nil {
		panic(err)
	}

	return &SynologyProvider{
		client:       client,
		DryRun:       providerConfig.DryRun,
		domainFilter: GetDomainFilter(*providerConfig),
	}
}

// Global functions

func GetDomainFilter(config Configuration) endpoint.DomainFilter {
	var domainFilter endpoint.DomainFilter
	createMsg := "Creating Synology provider with "

	if len(config.DomainList) > 0 {
		createMsg += fmt.Sprintf("zoneNode filter: '%s', ", strings.Join(config.DomainList, ","))
	}
	domainFilter = *endpoint.NewDomainFilter(config.DomainList)

	createMsg = strings.TrimSuffix(createMsg, ", ")
	if strings.HasSuffix(createMsg, "with ") {
		createMsg += "no kind of domain filters"
	}
	log.Info(createMsg)
	return domainFilter
}

// Functions called by the webhook http API

func (p *SynologyProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	records, err := p.client.RecordList(p.domainFilter.Filters, "master")
	if err != nil {
		return nil, err
	}

	var endpoints []*endpoint.Endpoint
	for _, r := range records {
		if provider.SupportedRecordType(string(r.Type)) && p.domainFilter.Match(r.Record) {
			endpoints = append(endpoints, endpoint.NewEndpoint(string(r.Record), string(r.Type), string(r.Value)))
		}
	}

	return endpoints, nil
}

func (p *SynologyProvider) AdjustEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	adjustedEndpoints := []*endpoint.Endpoint{}
	for _, ep := range endpoints {
		adjustedTargets := endpoint.Targets{}
		for _, t := range ep.Targets {
			err := p.client.RecordDelete(webapi.DNSRecord{
				Record: ep.DNSName,
				Type:   ep.RecordType,
				Value:  t,
			})
			if err != nil {
				log.Warning(err)
			}
			err = p.client.RecordCreate(webapi.DNSRecord{
				Record: ep.DNSName,
				Type:   ep.RecordType,
				Value:  t,
			})
			if err != nil {
				log.Warning(err)
			} else {
				adjustedTargets = append(adjustedTargets, t)
			}
		}
		ep.Targets = adjustedTargets
		adjustedEndpoints = append(adjustedEndpoints, ep)
	}
	return adjustedEndpoints, nil
}

func (p *SynologyProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	for _, ep := range append(changes.Delete, changes.UpdateOld...) {
		for _, t := range ep.Targets {
			err := p.client.RecordDelete(webapi.DNSRecord{
				Record: ep.DNSName,
				Type:   ep.RecordType,
				Value:  t,
			})
			if err != nil {
				log.Warning(err)
			}
		}
	}
	for _, ep := range append(changes.Create, changes.UpdateNew...) {
		for _, t := range ep.Targets {
			err := p.client.RecordCreate(webapi.DNSRecord{
				Record: ep.DNSName,
				Type:   ep.RecordType,
				Value:  t,
			})
			if err != nil {
				log.Warning(err)
			}
		}
	}
	return nil
}
