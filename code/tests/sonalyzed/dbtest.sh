#!/bin/bash
#
# You can run this test by creating a symlink data/fox.educloud.no to a directory which should
# contain information for June 2026 (ie a subdirectory "2026/06" with subdirectories "03" and "04"
# in the normal manner).  For various reasons this directory is not included in the sonalyze distro
# (too large; contains sensitive data).  It is included in the testdata/ directory of the
# sonar-deploy repo, internal to Sigma2/NRIS.  The checks below depend on the exact contents.

set -e

if [ ! -L data/fox.educloud.no ]; then
    echo '********************************************'
    echo "Skipping dbtest because no test data linked!"
    echo '********************************************'
    exit 0
fi

# Note if `sonalyze daemon` fails on startup the `set -e` will not catch it because the server is
# run in the background.  In this case, $sonalyzed_pid will reference a process that is not there.

testapi=127.0.0.1:4545
rm -rf $rootdir

# Set up a test jobanalyzer directory structure

# Run the server in the background against that directory

$SONALYZE daemon \
           -jobanalyzer-dir . \
           -rest-api $testapi \
           -v1 &
sonalyzed_pid=$!

# Always attempt to shut down the server on exit.  (Not sure if the HUP/INT are necessary or if they
# are subsumed by EXIT.)
trap "kill -HUP $sonalyzed_pid" EXIT ERR SIGHUP SIGINT

# Wait for sonalyzed to come up
sleep 1

output=$(curl --silent --fail-with-body -G "$testapi/api/v1/cards/fox.educloud.no?start_date=2026-06-03" |
             jq '.[]|.Model' | sort | uniq --count)
CHECK "dbtest_cards" \
      '     24 "NVIDIA A100 80GB PCIe"
     24 "NVIDIA A100-PCIE-40GB"
      8 "NVIDIA A40"
     32 "NVIDIA GeForce RTX 3090"
      8 "NVIDIA H100 NVL"
      4 "NVIDIA H100 PCIe"
     16 "NVIDIA H200 NVL"
     10 "NVIDIA L40S"' \
      "$output"

output=$(curl --silent --fail-with-body -G \
              "$testapi/api/v1/jobs/fox.educloud.no?start_date=2026-06-03&user=ec-eged,ec-markushs" |
             jq -r '.[]|.Cmd' | sort | uniq)
CHECK "dbtest_jobs" \
      "bedpostx_gpu,bedpostx_postpr,dwi_pre_tractog,post_proc_matri,python,python3.11,qunex,run_matrix1.sh,run_matrix3.sh,sh
bedpostx_gpu,bedpostx_postpr,post_proc_matri,python,python3.11,qunex,run_matrix1.sh,run_matrix3.sh,sh
bedpostx_gpu,make_trajectory,post_proc_matri,python,python3.11,qunex,run_matrix1.sh,run_matrix3.sh,sh
bedpostx_gpu,post_proc_matri,python,python3.11,qunex,run_matrix1.sh,run_matrix3.sh,sh
DeDriftAndResam,MCR Main thread,MSMAll.sh,ReApplyFixMulti,fslcpgeom,matlab_helper,python
DeDriftAndResam,MCR Main thread,MSMAll.sh,ReApplyFixMulti,matlab_helper,msm,python
DeDriftAndResam,MCR Main thread,MSMAll.sh,ReApplyFixMulti,matlab_helper,python
DeDriftAndResam,MCR Main thread,MSMAll.sh,ReApplyFixMulti,msm,python
DeDriftAndResam,MCR Main thread,MSMAll.sh,ReApplyFixMulti,python
DeDriftAndResam,MSMAll.sh,ReApplyFixMulti,fslcpgeom,matlab_helper,python
DeDriftAndResam,MSMAll.sh,ReApplyFixMulti,matlab_helper,python
DeDriftAndResam,MSMAll.sh,ReApplyFixMulti,python
DistortionCorre,GenericfMRISurf,GenericfMRIVolu,IntensityNormal,MCR Main thread,OneStepResampli,RibbonVolumeToS,epi_reg_dof,hcp_fix_multi_r,matlab_helper,python
DistortionCorre,GenericfMRISurf,GenericfMRIVolu,IntensityNormal,MCR Main thread,OneStepResampli,RibbonVolumeToS,fslmerge,hcp_fix_multi_r,matlab_helper,python
DistortionCorre,GenericfMRISurf,GenericfMRIVolu,MCR Main thread,MotionCorrectio,OneStepResampli,RibbonVolumeToS,epi_reg_dof,hcp_fix_multi_r,matlab_helper,mcflirt,python
DistortionCorre,GenericfMRISurf,GenericfMRIVolu,MCR Main thread,OneStepResampli,RibbonVolumeToS,convertwarp,fast,fslsplit,hcp_fix_multi_r,matlab_helper,mcflirt.sh,python,topup,wb_command
DistortionCorre,GenericfMRISurf,GenericfMRIVolu,MCR Main thread,OneStepResampli,RibbonVolumeToS,convertwarp,fast,hcp_fix_multi_r,matlab_helper,python
DistortionCorre,GenericfMRISurf,GenericfMRIVolu,MCR Main thread,OneStepResampli,RibbonVolumeToS,epi_reg_dof,fslmerge,fslsplit,hcp_fix_multi_r,matlab_helper,mcflirt,mcflirt.sh,python,topup
DistortionCorre,GenericfMRISurf,GenericfMRIVolu,MCR Main thread,OneStepResampli,RibbonVolumeToS,epi_reg_dof,hcp_fix_multi_r,matlab_helper,python
GenericfMRISurf,GenericfMRIVolu,IntensityNormal,MCR Main thread,OneStepResampli,RibbonVolumeToS,hcp_fix_multi_r,python
GenericfMRISurf,GenericfMRIVolu,MCR Main thread,OneStepResampli,RibbonVolumeToS,fslcpgeom,fslmaths,hcp_fix_multi_r,matlab_helper,melodic,python
GenericfMRISurf,GenericfMRIVolu,MCR Main thread,OneStepResampli,RibbonVolumeToS,fslmaths,hcp_fix_multi_r,python
GenericfMRISurf,GenericfMRIVolu,MCR Main thread,OneStepResampli,RibbonVolumeToS,hcp_fix_multi_r,matlab_helper,python
GenericfMRISurf,GenericfMRIVolu,MCR Main thread,OneStepResampli,RibbonVolumeToS,hcp_fix_multi_r,matlab_helper,python,squashfuse_ll
GenericfMRISurf,GenericfMRIVolu,MCR Main thread,RibbonVolumeToS,hcp_fix_multi_r,matlab_helper,python
GenericfMRISurf,MCR Main thread,hcp_fix_multi_r,matlab_helper,melodic,python,squashfuse_ll,wb_command
gzip,post_proc_matri,probtrackx2_gpu,python,qunex,run_matrix1.sh,run_matrix3.sh
gzip,post_proc_matri,probtrackx2_gpu,python,qunex,run_matrix1.sh,run_matrix3.sh,wb_command
gzip,slurm_script
post_proc_matri,python,python3.11,qunex,run_matrix1.sh,run_matrix3.sh,sh
post_proc_matri,python,python3.11,qunex,run_matrix1.sh,run_matrix3.sh,sh,xfibres_gpu10.2
post_proc_matri,python,qunex,run_matrix1.sh,run_matrix3.sh,wb_command
python
python3
python3,python3 <defunct>
python3,python3 <defunct>,srun
python3,srun
rapidtide,slurm_script
slurm_script
starter-suid
xfibres_gpu10.2" \
      "$output"

output=$(curl --silent --fail-with-body -G \
              "$testapi/api/v1/jobs/fox.educloud.no?start_date=2026-06-03&user=ec-emilp&fields=Hosts" |
             jq -r '.[]|.Hosts[]' | sort | uniq --count)
CHECK "dbtest_jobs_2" \
      '    108 c1-10
     51 c1-11
    159 c1-12
     75 c1-13
    123 c1-14
    114 c1-15
    145 c1-16
     89 c1-17
    112 c1-18
    133 c1-19
    135 c1-20
    135 c1-21
     50 c1-22
    124 c1-23
    103 c1-24
    103 c1-25
    126 c1-26
    205 c1-27
    118 c1-28
    142 c1-5
     99 c1-6
     98 c1-7
     76 c1-8
    221 c1-9' \
      "$output"

output=$(curl --silent --fail-with-body -G "$testapi/api/v1/nodes/fox.educloud.no?start_date=2026-06-03" |
             jq '.[]|.Description' | sort | uniq --count)
CHECK "dbtest_nodes" \
      '      4 "2x32 (hyperthreaded) AMD EPYC 7452 32-Core Processor, 503 GiB, 4x NVIDIA GeForce RTX 3090 @ 24GiB"
      2 "2x48 (hyperthreaded) AMD EPYC 7552 48-Core Processor, 1007 GiB, 4x NVIDIA A100-PCIE-40GB @ 40GiB"
      2 "2x48 (hyperthreaded) AMD EPYC 7642 48-Core Processor, 1007 GiB, 2x NVIDIA H100 PCIe @ 79GiB"
      6 "2x48 (hyperthreaded) AMD EPYC 7642 48-Core Processor, 1007 GiB, 4x NVIDIA A100 80GB PCIe @ 80GiB"
      4 "2x48 (hyperthreaded) AMD EPYC 7642 48-Core Processor, 1007 GiB, 4x NVIDIA A100-PCIE-40GB @ 40GiB"
      2 "2x48 (hyperthreaded) AMD EPYC 7642 48-Core Processor, 2003 GiB, 4x NVIDIA A40 @ 44GiB"
      2 "2x48 (hyperthreaded) AMD EPYC 7642 48-Core Processor, 2003 GiB, 8x NVIDIA GeForce RTX 3090 @ 24GiB"
      2 "2x64 AMD EPYC 7702 64-Core Processor, 1007 GiB"
     48 "2x64 AMD EPYC 7702 64-Core Processor, 503 GiB"
      2 "2x64 (hyperthreaded) AMD EPYC 7702 64-Core Processor, 2003 GiB"
      2 "2x64 (hyperthreaded) AMD EPYC 7713 64-Core Processor, 2003 GiB"
      2 "2x64 (hyperthreaded) AMD EPYC 7713 64-Core Processor, 2003 GiB, 1x NVIDIA L40S @ 44GiB"
      2 "2x64 (hyperthreaded) AMD EPYC 7H12 64-Core Processor, 1007 GiB, 4x NVIDIA L40S @ 44GiB"
      2 "2x64 (hyperthreaded) AMD EPYC 7H12 64-Core Processor, 2003 GiB, 4x NVIDIA H100 NVL @ 93GiB"
      2 "2x96 (hyperthreaded) AMD EPYC 9655 96-Core Processor, 1511 GiB, 8x NVIDIA H200 NVL @ 140GiB"' \
          "$output"

