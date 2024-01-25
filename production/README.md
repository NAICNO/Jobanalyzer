# Sonar and Jobanalyzer production setups

## Cluster names

Every cluster has a *cluster name* that distinguishes it globally.  Frequently the cluster name is
the FQDN of the cluster's login node.  These clusters are defined at present:

* `mlx.hpc.uio.no` - abbreviated `ml` and `mlx` - UiO machine learning nodes
* `fox.educloud.no` - abbreviated `fox` - UiO "Fox" supercomputer
* `saga.sigma2.no` - abbreviated `saga` - Sigma2/NRIS "Saga" supercomputer

The cluster name scheme is imperfectly implemented throughout the system but we are moving in the
direction of using it for everything.

Clusters have individual setups, owing partly to how the setups have evolved independently and
partly to real differences between the clusters.  There are however many commonalities.


## Compute nodes, analysis nodes, web nodes

On every *compute node* in a cluster we run the analysis program `sonar` to obtain *sample data*.
The sample data are exfiltrated to the *analysis node* where the data are aggregated and analyzed.
The analyses generate *reports* that are uploaded to the *web node*.  The web node serves the
*dashboard* which delivers the reports to web clients that request them.  The web node can also
serve interactive queries, which it runs by contacting the analysis node.

Thus to set up Jobanalyzer, one must set up `sonar` on each of the compute nodes, the data
management and analysis framework on the analysis node, and the web server and dashboard framework
on the web node.

Everything pertaining to the *compute nodes* is in the subdirectory `sonar-nodes`.  See the README in
that directory for how to set up `sonar` on compute nodes.

At this time, the analysis node and web nodes are the same node (and it will probably remain that
way).  Everything pertaining to this joint node is in the subdirectory `jobanalyzer-server`.  See
the README in that directory for how to set up analysis and web service on the server.
