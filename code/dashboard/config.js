// Dashboard configuration data.

// Set this to true to load mocked-up data everywhere
var TESTDATA = false;

// Cluster-info returns information about the cluster.  This function must be manually updated
// whenever we add a cluster.

function cluster_info(cluster) {
    switch (cluster) {
    default:
        /*FALLTHROUGH*/
    case "ml":
        return {
            cluster,
            subclusters: ["nvidia"],
            uptime: true,
            violators: true,
            deadweight: true,
            defaultQuery: "*",
            hasDowntime: true,
            crossNode: false,
            name:"ML nodes",
            description:"UiO Machine Learning nodes",
            prefix:"ml-",
            policy:"Significant CPU usage without any GPU usage",
        }
    case "fox":
        return {
            cluster,
            subclusters: ["cpu","gpu","int","login"],
            uptime: true,
            violators: false,
            deadweight: true,
            defaultQuery: "login* or int*",
            name:"Fox",
            hasDowntime: true,
            crossNode: true,
            description:"UiO 'Fox' supercomputer",
            prefix:"fox-",
            policy:"(To be determined)",
        }
    case "saga":
        return {
            cluster,
            subclusters: ["login"],
            uptime: false,
            violators: false,
            deadweight: false,
            defaultQuery: "login*",
            name:"Saga",
            hasDowntime: false,
            crossNode: true,
            description:"Sigma2 'Saga' supercomputer",
            prefix:"saga-",
            policy:"(To be determined)",
        }
    }
}
