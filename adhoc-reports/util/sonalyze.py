# Run sonalyze and return its stdout as a list of lines
#
# TODO: Some things could be keyword arguments
#
# TODO: Whether to print or return the output could be an option

import os, subprocess

def sonalyze(verb, cluster, args):
    sonalyze_bin=os.path.join(os.getenv("HOME"), "go/bin/sonalyze")
    remote = "https://naic-monitor.uio.no"
    auth_file = os.path.join(os.getenv("HOME"), ".ssh/sonalyzed-auth.txt")

    cmd = [sonalyze_bin,
            verb,
            "-remote", remote,
            "-auth-file", auth_file,
            "-cluster", cluster,
            *args]
    proc = subprocess.run(cmd, capture_output=True, text=True, cwd=os.getcwd())
    if proc.returncode != 0:
        raise RuntimeError("sonalyze " + verb + " failed:" + proc.stderr)
    return proc.stdout.splitlines()

