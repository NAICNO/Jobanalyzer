interface HostDetails {
  hostname: string;
  date: Date;
  tag: string;
  bucketing: string;
  labels: string[];
  rcpu: number[];
  rmem: number[];
  rres: number[];
  rgpu?: number[];
  rgpumem?: number[];
  downhost?: 0 | 1;
  downgpu?: 0 | 1;
  system: {
    hostname: string;
    description: string;
  }
}
