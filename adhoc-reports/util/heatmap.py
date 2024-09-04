# Run sonalyze and return its stdout as a list of lines
#
# TODO: Whether to print or return the output could be an option

import os, subprocess

# data is a list of 2-element lists, the values must be convertible with str()
def heatmap(data, gridsize):
    heatmap_bin=os.path.join(os.getenv("HOME"), "go/bin/heatmap")
    s = ""
    for x in data:
        if s != "":
            s += "\n"
        s += str(x[0]) + " " + str(x[1])
    proc = subprocess.run([heatmap_bin,
                              "-n", str(gridsize)],
                             input=s,
                             text=True,
                             capture_output=True,
                             cwd=os.getcwd())
    if proc.returncode != 0:
        raise RuntimeError("heatmap failed:" + proc.stderr)
    return proc.stdout.splitlines()

