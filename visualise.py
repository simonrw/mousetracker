#!/usr/bin/env python

import argparse
import sqlite3
from datetime import datetime, timezone
import logging
from pathlib import Path
import pandas as pd
import matplotlib.pyplot as plt

logging.basicConfig(level=logging.WARNING, format="[%(asctime)s] %(message)s")
LOG = logging.getLogger(__name__)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-d",
        "--db",
        required=False,
        default=Path.home().joinpath(".local", "share", "mousetracker", "db.db"),
        type=Path,
    )
    parser.add_argument("-v", "--verbose", action="count", default=0)
    args = parser.parse_args()

    if args.verbose == 1:
        LOG.setLevel(logging.INFO)
    elif args.verbose > 1:
        LOG.setLevel(logging.DEBUG)

    LOG.info("loading state from %s", args.db)

    conn = sqlite3.connect(args.db)
    start, end = conn.execute("select min(start), max(end) from sessions").fetchone()
    start = datetime.strptime(start, "%Y-%m-%d %H:%M:%S.%f%z")
    end = datetime.strptime(end, "%Y-%m-%d %H:%M:%S.%f%z")

    df = pd.read_sql_query(
        "SELECT * FROM sessions",
        conn,
    )

    LOG.info("calculating")

    # fraction of the day spent moving the mouse
    total_time_seconds = 0
    for _, session in df.iterrows():
        duration = pd.to_datetime(session.end) - pd.to_datetime(session.start)
        total_time_seconds += duration.total_seconds()

    print(f"Total days: {end - start}")
    print(f"Total mouse seconds: {total_time_seconds}")
    fraction = total_time_seconds * 100.0 / (end - start).total_seconds()
    print(f"Fraction of mouse movements: {fraction:.2f} %")
