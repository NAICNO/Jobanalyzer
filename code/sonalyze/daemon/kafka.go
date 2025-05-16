package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"github.com/twmb/franz-go/pkg/kgo"

	"sonalyze/db"
)

const (
	tySample = iota
	tySysinfo
	tyJobs
	tyCluster
)

// This runs on a goroutine - one goroutine per cluster, just to be a little resilient.

func runKafka(kafkaBroker, cluster string, ds *db.PersistentCluster, verbose bool) {
	defer ds.Close()
	handler := newClusterHandler(cluster, ds, verbose)
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(kafkaBroker),
		kgo.ConsumerGroup("jobanalyzer-ingest"),
		kgo.ConsumeTopics(
			handler.add(tySample),
			handler.add(tySysinfo),
			handler.add(tyJobs),
			handler.add(tyCluster),
		))
	if err != nil {
		// This should be surfaced somehow, but probably we should just back off and retry later,
		// the broker could be down - depends on the error!
		log.Printf("%s: Failed to create client: %v", cluster, err)
		return
	}
	defer cl.Close()
	if verbose {
		log.Printf("%s: Connected!", cluster)
	}

	ctx := context.Background()
	for {
		if verbose {
			log.Printf("%s: Fetching data", cluster)
		}
		fetches := cl.PollFetches(ctx)
		if verbose {
			log.Printf("%s: Fetched data", cluster)
		}
		if errs := fetches.Errors(); len(errs) > 0 {
			// All errors are retried internally when fetching, but non-retriable errors are
			// returned from polls so that users can notice and take action.
			log.Printf("%s: SOFT ERROR: Failed to fetch data! %v", cluster, errs)
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			if verbose {
				log.Printf("  %s: %s", cluster, record.Topic)
			}
			err := handler.dispatch(record.Topic, record.Key, record.Value)
			if err != nil {
				log.Printf("  %s: SOFT ERROR: Topic handler %s failed: %v", cluster, record.Topic, err)
			}
		}
		if err := cl.CommitUncommittedOffsets(ctx); err != nil {
			log.Printf("  %s: SOFT ERROR: Commit records failed: %v", cluster, err)
		}
	}
}

type clusterHandler struct {
	cluster string
	disp    map[string]func(ch *clusterHandler, topic, host string, data []byte) error
	ds      *db.PersistentCluster
	verbose bool
}

func newClusterHandler(cluster string, ds *db.PersistentCluster, verbose bool) *clusterHandler {
	return &clusterHandler{
		cluster: cluster,
		disp:    make(map[string]func(ch *clusterHandler, topic, host string, data []byte) error),
		ds:      ds,
		verbose: verbose,
	}
}

func (ch *clusterHandler) add(ty int) string {
	var name string
	switch ty {
	case tySample:
		name = ch.cluster + "." + string(newfmt.DataTagSample)
		ch.disp[name] = handleSample
	case tySysinfo:
		name = ch.cluster + "." + string(newfmt.DataTagSysinfo)
		ch.disp[name] = handleSysinfo
	case tyJobs:
		name = ch.cluster + "." + string(newfmt.DataTagJobs)
		ch.disp[name] = handleSlurmJobs
	case tyCluster:
		name = ch.cluster + "." + string(newfmt.DataTagCluster)
		ch.disp[name] = handleCluster
	default:
		panic("No such type")
	}
	return name
}

func (ch *clusterHandler) dispatch(topic string, key, value []byte) error {
	if handler, found := ch.disp[topic]; found {
		defer ch.ds.FlushAsync()
		return handler(ch, topic, string(key), value)
	}
	return fmt.Errorf("%s: No handler for topic: %s", ch.cluster, topic)
}

func handleSample(ch *clusterHandler, topic, host string, data []byte) error {
	info := new(newfmt.SampleEnvelope)
	err := json.Unmarshal(data, info)
	if err != nil {
		return err
	}
	if info.Data != nil {
		if ch.verbose {
			log.Printf("%s: Got a good sample %s %s", ch.cluster, topic, host)
		}
		return ch.ds.AppendSamplesAsync(db.FileSampleV0JSON, host, string(info.Data.Attributes.Time), data)
	}
	if ch.verbose {
		log.Printf("%s: Dropping a sample error object on the floor", ch.cluster)
	}
	return nil
}

func handleSysinfo(ch *clusterHandler, topic, host string, data []byte) error {
	info := new(newfmt.SysinfoEnvelope)
	err := json.Unmarshal(data, info)
	if err != nil {
		return err
	}
	if info.Data != nil {
		if ch.verbose {
			log.Printf("%s: Got a good sysinfo %s %s", ch.cluster, topic, host)
		}
		return ch.ds.AppendSysinfoAsync(db.FileSysinfoV0JSON, host, string(info.Data.Attributes.Time), data)
	}
	if ch.verbose {
		log.Printf("%s: Dropping a sysinfo error object on the floor", ch.cluster)
	}
	return nil
}

func handleSlurmJobs(ch *clusterHandler, topic, host string, data []byte) error {
	info := new(newfmt.JobsEnvelope)
	err := json.Unmarshal(data, info)
	if err != nil {
		return err
	}
	if info.Data != nil {
		if ch.verbose {
			log.Printf("%s: Got a good jobs %s %s", ch.cluster, topic, host)
		}
		return ch.ds.AppendSlurmSacctAsync(db.FileSlurmV0JSON, string(info.Data.Attributes.Time), data)
	}
	if ch.verbose {
		log.Printf("%s: Dropping a job error object on the floor", ch.cluster)
	}
	return nil
}

func handleCluster(ch *clusterHandler, topic, host string, data []byte) error {
	info := new(newfmt.ClusterEnvelope)
	err := json.Unmarshal(data, info)
	if err != nil {
		return err
	}
	if info.Data != nil {
		if ch.verbose {
			log.Printf("%s: Got a good cluster %s %s", ch.cluster, topic, host)
		}
		return ch.ds.AppendCluzterAsync(db.FileCluzterV0JSON, string(info.Data.Attributes.Time), data)
	}
	if ch.verbose {
		log.Printf("%s: Dropping a cluster error object on the floor", ch.cluster)
	}
	return nil
}
