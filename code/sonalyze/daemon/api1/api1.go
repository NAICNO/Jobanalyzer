// The v1 API follows the old v0 API but GET requests returns proper JSON objects instead of strings, and the JSON
// objects can have non-string values.
//
// The v1 API also has new insertion points for the new data, represented as JSON.  The result of a
// POST is a JSON object with some data about the data that were received.

package api1

import (
	"context"
	"encoding/json"

	"go-utils/auth"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"github.com/danielgtaylor/huma/v2"

	"sonalyze/daemon/apiutil"
	"sonalyze/db"
)

var (
	postAuthenticator *auth.Authenticator
)

func SetupAPI(
	api huma.API,
	insertAPI bool,
	postAuthenticator_ *auth.Authenticator,
) {
	postAuthenticator = postAuthenticator_
	grp := huma.NewGroup(api, "/api/v1")

	// v1 get apis go here, when we've implemented them

	if insertAPI {
		addInsertSysinfoData(grp)
		addInsertSampleData(grp)
		addInsertJobData(grp)
		addInsertClusterData(grp)
	}
}

// Insertion.
//
// Sonar does not require a specific return structure beyond the HTTP code.  Here, on successful
// insertion, echo the cluster/node/topic/time back, since sonar assumes these will be unique.
//
// Insertion ops must return `error` to be API compatible with Huma, but the error return is always
// a huma.StatusError.

type InsertionResponse struct {
	Body InsertionResponseBody
}

type InsertionResponseBody struct {
	Cluster string `json:"cluster"`
	Node    string `json:"node,omitempty"` // There's no node for jobs and cluster data
	Topic   string `json:"topic"`
	Time    string `json:"time"`
}

const (
	insertSampleName  = "/insert/" + string(newfmt.DataTagSample)
	insertSysinfoName = "/insert/" + string(newfmt.DataTagSysinfo)
	insertJobsName    = "/insert/" + string(newfmt.DataTagJobs)
	insertClusterName = "/insert/" + string(newfmt.DataTagCluster)
)

func addInsertSysinfoData(api huma.API) {
	huma.Post(api, insertSysinfoName, func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			Body newfmt.SysinfoEnvelope
		},
	) (*InsertionResponse, error) {
		cluster := string(input.Body.Data.Attributes.Cluster)
		ds, hErr := insertionSetup(insertSysinfoName, cluster, input.Auth)
		if hErr != nil {
			return nil, hErr
		}
		defer ds.FlushAsync()
		nodename := string(input.Body.Data.Attributes.Node)
		timestamp := string(input.Body.Data.Attributes.Time)
		payload, _ := json.Marshal(input.Body)
		err := ds.AppendSysinfoAsync(db.DataSysinfoV0JSON, nodename, timestamp, payload)
		if err != nil {
			return nil, huma.Error400BadRequest("insert: " + err.Error())
		}
		return insertionResponse(cluster, nodename, timestamp, newfmt.DataTagSysinfo), nil
	})
}

func addInsertSampleData(api huma.API) {
	huma.Post(api, insertSampleName, func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			Body newfmt.SampleEnvelope
		},
	) (*InsertionResponse, error) {
		cluster := string(input.Body.Data.Attributes.Cluster)
		ds, hErr := insertionSetup(insertSampleName, cluster, input.Auth)
		if hErr != nil {
			return nil, hErr
		}
		defer ds.FlushAsync()
		nodename := string(input.Body.Data.Attributes.Node)
		timestamp := string(input.Body.Data.Attributes.Time)
		payload, _ := json.Marshal(input.Body)
		err := ds.AppendSamplesAsync(db.DataSampleV0JSON, nodename, timestamp, payload)
		if err != nil {
			return nil, huma.Error400BadRequest("insert: " + err.Error())
		}
		return insertionResponse(cluster, nodename, timestamp, newfmt.DataTagSample), nil
	})
}

func addInsertJobData(api huma.API) {
	huma.Post(api, insertJobsName, func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			Body newfmt.JobsEnvelope
		},
	) (*InsertionResponse, error) {
		cluster := string(input.Body.Data.Attributes.Cluster)
		ds, hErr := insertionSetup(insertJobsName, cluster, input.Auth)
		if hErr != nil {
			return nil, hErr
		}
		defer ds.FlushAsync()
		timestamp := string(input.Body.Data.Attributes.Time)
		payload, _ := json.Marshal(input.Body)
		err := ds.AppendSlurmSacctAsync(db.DataSlurmV0JSON, timestamp, payload)
		if err != nil {
			return nil, huma.Error400BadRequest("insert: " + err.Error())
		}
		return insertionResponse(cluster, "", timestamp, newfmt.DataTagJobs), nil
	})
}

func addInsertClusterData(api huma.API) {
	huma.Post(api, insertClusterName, func(
		ctx context.Context,
		input *struct {
			apiutil.AuthHeader
			Body newfmt.ClusterEnvelope
		},
	) (*InsertionResponse, error) {
		cluster := string(input.Body.Data.Attributes.Cluster)
		ds, hErr := insertionSetup(insertClusterName, cluster, input.Auth)
		if hErr != nil {
			return nil, hErr
		}
		defer ds.FlushAsync()
		timestamp := string(input.Body.Data.Attributes.Time)
		payload, _ := json.Marshal(input.Body)
		err := ds.AppendCluzterAsync(db.DataCluzterV0JSON, timestamp, payload)
		if err != nil {
			return nil, huma.Error400BadRequest("insert: " + err.Error())
		}
		return insertionResponse(cluster, "", timestamp, newfmt.DataTagCluster), nil
	})
}

func insertionResponse(
	cluster, nodename, timestamp string,
	datatype newfmt.DataType,
) *InsertionResponse {
	return &InsertionResponse{
		Body: InsertionResponseBody{
			Cluster: cluster,
			Node:    nodename,
			Topic:   string(newfmt.DataTagSysinfo),
			Time:    timestamp,
		},
	}
}

func insertionSetup(path, cluster, auth string) (db.AppendablePersistentDataProvider, huma.StatusError) {
	if postAuthenticator != nil {
		user, pass := apiutil.DecodeAuth(auth)
		if user != cluster {
			return nil, huma.Error401Unauthorized("insert: Cluster in data does not match user in auth")
		}
		if !postAuthenticator.Authenticate(user, pass) {
			return nil, huma.Error401Unauthorized("insert: Unknown user/pass combination")
		}
	}
	meta, hErr := apiutil.GetClusterContext(insertSysinfoName, cluster)
	if hErr != nil {
		return nil, hErr
	}
	ds, err := db.OpenAppendablePersistentDirectoryDB(meta)
	if err != nil {
		return nil, huma.Error500InternalServerError("insert: incompatible database")
	}
	return ds, nil
}
