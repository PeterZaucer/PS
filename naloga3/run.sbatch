#!/bin/bash
#SBATCH --nodes=1
#SBATCH --array=0-20
#SBATCH --reservation=fri
#SBATCH --output=telefon-%a.txt

module load Go
srun go run naloga3.go -port 9000 -id $SLURM_ARRAY_TASK_ID -n $SLURM_ARRAY_TASK_COUNT
~/go/bin/GoVector --log_type shiviz --log_dir . --outfile Log-Full.log