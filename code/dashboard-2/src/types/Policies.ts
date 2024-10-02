export interface Policy {
  name: string;
  trigger: string;
  problem: string;
  remedy: string;
}

export interface Policies {
  [clusterName: string]: Policy[];
}
