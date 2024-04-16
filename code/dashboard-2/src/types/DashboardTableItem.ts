interface DashboardTableItem {
  hostname: string;
  tag: string;
  machine: string;
  recent: number;
  longer: number;
  long: number;
  cpu_status: number;
  gpu_status: number;
  jobs_recent: number;
  jobs_longer: number;
  users_recent: number;
  users_longer: number;
  cpu_recent: number;
  cpu_longer: number;
  mem_recent: number;
  mem_longer: number;
  resident_recent: number;
  resident_longer: number;
  gpu_recent: number;
  gpu_longer: number;
  gpumem_recent: number;
  gpumem_longer: number;
  violators_long: number;
  zombies_long: number;
}
