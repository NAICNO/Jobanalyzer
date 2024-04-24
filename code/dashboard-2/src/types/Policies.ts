interface Policy {
  name: string;
  trigger: string;
  problem: string;
  remedy: string;
}

interface Policies {
  [clusterName: string]: Policy[];
}
