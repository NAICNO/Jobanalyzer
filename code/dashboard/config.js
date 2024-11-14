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
            canonical: "mlx.hpc.uio.no",
            subclusters: [{name:"nvidia", nodes:"ml[1-3,6-9]"}],
            uptime: true,
            violators: false,
            deadweight: false,
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
            canonical: "fox.educloud.no",
            subclusters: [{name:"cpu", nodes:"c*"},
                          {name:"gpu", nodes:"gpu*"},
                          {name:"int", nodes:"int*"},
                          {name:"login", nodes:"login*"}],
            uptime: true,
            violators: false,
            deadweight: false,
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
            canonical: "saga.sigma2.no",
            subclusters: [],
            uptime: false,
            violators: false,
            deadweight: false,
            defaultQuery: "c*-1",
            name:"Saga",
            hasDowntime: false,
            description:"Sigma2 'Saga' supercomputer",
            prefix:"saga-",
            policy:"(To be determined)",
        }
    case "fram":
        return {
            cluster,
            canonical: "fram.sigma2.no",
            subclusters: [],
            uptime: false,
            violators: false,
            deadweight: false,
            defaultQuery: "c*-1",
            name:"Fram",
            hasDowntime: false,
            description:"Sigma2 'Fram' supercomputer",
            prefix:"fram-",
            policy:"(To be determined)",
        }
    case "betzy":
        return {
            cluster,
            canonical: "betzy.sigma2.no",
            subclusters: [],
            uptime: false,
            violators: false,
            deadweight: false,
            defaultQuery: "b11*",
            name:"Betzy",
            hasDowntime: false,
            description:"Sigma2 'Betzy' supercomputer",
            prefix:"betzy-",
            policy:"(To be determined)",
        }
    }
}
