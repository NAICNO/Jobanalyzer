#!/bin/bash
rsync -e "ssh -i $HOME/sonar/secrets/saga_rsync" -rt larstha@saga.sigma2.no:/cluster/shared/sonar/data/2024 $HOME/sonar/data/saga.sigma2.no
