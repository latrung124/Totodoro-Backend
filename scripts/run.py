#!/usr/bin/env python3
import argparse
import os
import platform
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SCRIPTS = ROOT / "scripts"

def is_windows() -> bool:
    return os.name == "nt" or platform.system().lower() == "windows"

def script_path(kind: str) -> Path:
    if kind not in ("setup", "build"):
        raise ValueError("kind must be 'setup' or 'build'")
    if is_windows():
        name = f"{kind}-windows.bat"
    else:
        # Reuse ubuntu script for Linux/macOS; note: macOS parts requiring apt-get will be skipped/fail
        name = f"{kind}-ubuntu.sh"
    return SCRIPTS / name

def ensure_executable(p: Path) -> None:
    if is_windows():
        return
    try:
        mode = p.stat().st_mode
        p.chmod(mode | 0o111)
    except Exception:
        pass

def run(kind: str, extra_args: list[str]) -> int:
    sp = script_path(kind)
    if not sp.exists():
        print(f"Error: script not found: {sp}", file=sys.stderr)
        return 1

    ensure_executable(sp)

    print(f"Detected OS: {platform.system()}  -> running: {sp}")
    cwd = str(ROOT)

    if is_windows():
        # Use cmd to run .bat; pass args as a single command line
        cmd = ["cmd.exe", "/c", str(sp)]
        if extra_args:
            cmd.extend(extra_args)
        return subprocess.call(cmd, cwd=cwd)
    else:
        # Prefer bash if available; fallback to sh
        sh = shutil.which("bash") or shutil.which("sh")
        if not sh:
            print("Error: neither bash nor sh found on PATH", file=sys.stderr)
            return 1
        cmd = [sh, str(sp)]
        if extra_args:
            cmd.extend(extra_args)
        return subprocess.call(cmd, cwd=cwd)

def main() -> None:
    parser = argparse.ArgumentParser(description="Run setup/build with OS-specific scripts.")
    parser.add_argument("action", choices=["setup", "build"], help="Which action to run")
    parser.add_argument("args", nargs=argparse.REMAINDER, help="Extra args passed to the underlying script")
    ns = parser.parse_args()

    # Strip leading "--" separator if used (common when forwarding args)
    extra = ns.args
    if extra and extra[0] == "--":
        extra = extra[1:]

    code = run(ns.action, extra)
    sys.exit(code)

if __name__ == "__main__":
    main()