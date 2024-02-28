// Dashboard configuration data.

// Set this to true to load mocked-up data everywhere
var TESTDATA = false;

// Cluster-info returns information about the cluster.  This function must be manually updated
// whenever we add a cluster or subcluster.  THIS IS BRITTLE.  Bug #379 wants to fix this,
// eventually.

function cluster_info(cluster) {
    switch (cluster) {
    default:
        /*FALLTHROUGH*/
    case "ml":
        return {
            cluster,
            subclusters: [{name:"nvidia", nodes:"ml[1-3,6-9]"}],
            uptime: true,
            violators: true,
            deadweight: true,
            defaultQuery: "*",
            hasDowntime: true,
            name:"ML nodes",
            description:"UiO Machine Learning nodes",
            prefix:"ml-",
            policy:"Significant CPU usage without any GPU usage",
        }
    case "fox":
        return {
            cluster,
            subclusters: [{name:"cpu", nodes:"c*"},
                          {name:"gpu", nodes:"gpu*"},
                          {name:"int", nodes:"int*"},
                          {name:"login", nodes:"login*"}],
            uptime: true,
            violators: false,
            deadweight: true,
            defaultQuery: "login* or int*",
            name:"Fox",
            hasDowntime: true,
            description:"UiO 'Fox' supercomputer",
            prefix:"fox-",
            policy:"(To be determined)",
        }
    case "saga":
        return {
            cluster,
            subclusters: [{name:"login", nodes:"login*"}],
            uptime: false,
            violators: false,
            deadweight: false,
            defaultQuery: "login*",
            name:"Saga",
            hasDowntime: false,
            description:"Sigma2 'Saga' supercomputer",
            prefix:"saga-",
            policy:"(To be determined)",
        }
    }
}
