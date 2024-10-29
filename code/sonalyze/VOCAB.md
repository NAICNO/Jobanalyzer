# Data field vocabulary

(This is evolving.)

Field naming is pretty arbitrary and it is not going to be cleaned up right now.  For the most part
we can fix things over time through the use of aliases.

"Old" names such as "rcpu", "rmem" should probably not be used more than absolutely necessary,
ideally all new names are fairly self-explanatory and not very abbreviated.

Contextuality is important to make things hang together.  The precise meaning of the field must be
derivable from name + context + type + documentation, ideally from name + context + documentation
since the user may not have access to the type.  Name + documentation must be visible from -fmt
help, and context is given by the verb.  (Hence plain "Name" in the cluster info is not as bad as it
looks because it is plain from context and documentation that we're talking about the cluster name;
"Clustername" might have been better, but not massively much better.)

Spelling standards that we should follow when we have a chance to (re)name a field:

* Cpu, Cpus not CPU, CPUS, CPUs
* Gpu, Gpus not GPU, GPUS, GPUs
* GB not GiB, the unit is 2^30
* MB not MiB, the unit is 2^20
* KB not KiB, the unit is 2^10
* JobId not JobID
* Units on fields that can have multiple natural units, eg, ResidentMemGB not ResidentMem

(And yet there may be other considerations.  The sacct table names such as UsedCPU and MaxRSS are
currently the way they are because those are the names adopted by sacct.  But on the whole it'd
probably be better to follow our own naming standards and explain the mapping in the documentation.)
