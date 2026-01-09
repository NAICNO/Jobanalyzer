/**
 * Job state constants based on SLURM job states
 * @see https://slurm.schedmd.com/job_state_codes.html#states
 */
export const JobState = {
  // Base job states
  BOOT_FAIL: 'BOOT_FAIL',
  CANCELLED: 'CANCELLED',
  COMPLETED: 'COMPLETED',
  DEADLINE: 'DEADLINE',
  FAILED: 'FAILED',
  NODE_FAIL: 'NODE_FAIL',
  OUT_OF_MEMORY: 'OUT_OF_MEMORY',
  PENDING: 'PENDING',
  PREEMPTED: 'PREEMPTED',
  RUNNING: 'RUNNING',
  SUSPENDED: 'SUSPENDED',
  TIMEOUT: 'TIMEOUT',
} as const

/**
 * Job flags that may be set in addition to base states
 * @see https://slurm.schedmd.com/job_state_codes.html#flags
 */
export const JobFlag = {
  COMPLETING: 'COMPLETING',
  CONFIGURING: 'CONFIGURING',
  EXPEDITING: 'EXPEDITING',
  LAUNCH_FAILED: 'LAUNCH_FAILED',
  POWER_UP_NODE: 'POWER_UP_NODE',
  RECONFIG_FAIL: 'RECONFIG_FAIL',
  REQUEUED: 'REQUEUED',
  REQUEUE_FED: 'REQUEUE_FED',
  REQUEUE_HOLD: 'REQUEUE_HOLD',
  RESIZING: 'RESIZING',
  RESV_DEL_HOLD: 'RESV_DEL_HOLD',
  REVOKED: 'REVOKED',
  SIGNALING: 'SIGNALING',
  SPECIAL_EXIT: 'SPECIAL_EXIT',
  STAGE_OUT: 'STAGE_OUT',
  STOPPED: 'STOPPED',
  UPDATE_DB: 'UPDATE_DB',
} as const

export type JobStateType = typeof JobState[keyof typeof JobState]
export type JobFlagType = typeof JobFlag[keyof typeof JobFlag]

/**
 * List of all job states for UI components (dropdowns, filters, etc.)
 */
export const JOB_STATES = Object.values(JobState)

/**
 * List of all job flags
 */
export const JOB_FLAGS = Object.values(JobFlag)
