import datetime as dt
import logging
import os
import sqlite3
import time
import uuid
from collections import namedtuple
from typing import List

from dotenv import load_dotenv
from skyfield.api import EarthSatellite, load
from skyfield.toposlib import wgs84

Notification = namedtuple("Notification", ["Service"])
Station = namedtuple(
    "Station",
    ["StnID", "StnName", "Latitude", "Longitude", "Altitude", "MinHorizon", "Status"],
)
Satellite = namedtuple("Satellite", ["NoradID", "SatName", "Status", "Line1", "Line2"])
Pass = namedtuple(
    "Pass",
    [
        "PassID",
        "StnID",
        "StnName",
        "NoradID",
        "SatName",
        "Azimuth",
        "Elevation",
        "AOS",
        "LOS",
    ],
)

logger = logging.getLogger(__name__)


def QuerySats(cursor: sqlite3.Cursor) -> List[Satellite]:
    query_satellites = """
        SELECT
            noradid,
            satname,
            status,
            line1,
            line2
        FROM Satellites
        WHERE status='online';
    """

    sat_rows = cursor.execute(query_satellites)
    sats = []

    for sat in sat_rows.fetchall():
        sats.append(Satellite(sat[0], sat[1], sat[2], sat[3], sat[4]))

    return sats


def QueryStns(cursor: sqlite3.Cursor) -> List[Station]:
    query_stations = """
        SELECT
            stnid,
            stnname,
            latitude,
            longitude,
            altitude,
            minhorizon,
            status
        FROM Stations
        WHERE status='online';
    """

    stn_rows = cursor.execute(query_stations)
    stns = []

    for stn in stn_rows.fetchall():
        stns.append(Station(stn[0], stn[1], stn[2], stn[3], stn[4], stn[5], stn[6]))

    return stns


def ComputePasses(
    stns: List[Station], sats: List[Satellite], max_horizon: int
) -> List[Pass]:
    passes = []

    for stn in stns:
        stn_wgs84 = wgs84.latlon(stn.Latitude, stn.Longitude)

        for sat in sats:
            earth_sat = EarthSatellite(
                sat.Line1, sat.Line2, name=sat.SatName, ts=load.timescale()
            )
            start = load.timescale().now()
            end = start + dt.timedelta(hours=max_horizon)
            t, events = earth_sat.find_events(
                stn_wgs84, start, end, altitude_degrees=stn.MinHorizon
            )
            starts, ends = [], []

            for ti, event in zip(t, events):
                match event:
                    case 0:
                        starts.append(ti.utc_datetime())

                    case 1:
                        continue

                    case 2:
                        ends.append(ti.utc_datetime())

            for aos, los in zip(starts, ends):
                diff = earth_sat - stn_wgs84
                ts = load.timescale().from_datetime(aos)
                topo_pos = diff.at(ts)
                ele, azi, _ = topo_pos.altaz()
                ps = Pass(
                    str(uuid.uuid4()),
                    stn.StnID,
                    stn.StnName,
                    sat.NoradID,
                    sat.SatName,
                    float(azi.degrees),
                    float(ele.degrees),
                    aos.isoformat(),
                    los.isoformat(),
                )
                passes.append(ps)

    return passes


def InsertPasses(
    conn: sqlite3.Connection, cursor: sqlite3.Cursor, passes: List[Pass]
) -> None:
    insert_pass = """
        INSERT INTO Passes (
            passid,
            stnid,
            stnname,
            noradid,
            satname,
            azimuth,
            elevation,
            aos,
            los
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
    """

    try:
        cursor.execute("BEGIN")
        for ps in passes:
            cursor.execute(insert_pass, ps)
        conn.commit()

    except Exception as e:
        conn.rollback()
        logger.error(f"Error: {e}")

    return


def SelectMaxHorizon(cursor: sqlite3.Cursor) -> int:
    horizon = """
        SELECT
            MAX(max_horizon)
        FROM Parameters;
    """
    cursor.execute(horizon)
    row = cursor.fetchone()
    if row is None:
        return 24

    max_horizon = int(row[0])
    return max_horizon


def DeletePasses(conn: sqlite3.Connection, cursor: sqlite3.Cursor) -> None:
    delete_passes = """
        DELETE
        FROM Passes;
    """

    try:
        cursor.execute("BEGIN")
        cursor.execute(delete_passes)
        conn.commit()

    except Exception as e:
        conn.rollback()
        logger.error(f"Error: {e}")

    return


def DeleteHorizons(conn: sqlite3.Connection, cursor: sqlite3.Cursor) -> None:
    delete_horizons = """
        DELETE
        FROM Parameters
        WHERE max_horizon IS NOT NULL;
    """

    try:
        cursor.execute("BEGIN")
        cursor.execute(delete_horizons)
        conn.commit()

    except Exception as e:
        conn.rollback()
        logger.error(f"Error: {e}")

    return


def DeleteOrbitalNotifs(conn: sqlite3.Connection, cursor: sqlite3.Cursor) -> None:
    delete_notifs = """
        DELETE
        FROM Notifications
        WHERE service='orbital';
    """

    try:
        cursor.execute("BEGIN")
        cursor.execute(delete_notifs)
        conn.commit()

    except Exception as e:
        conn.rollback()
        logger.error(f"Error: {e}")

    return


def CountNotifs(cursor: sqlite3.Cursor) -> int:
    count_notifs = """
        SELECT
            COUNT(service)
        FROM Notifications
        WHERE service='orbital';
    """
    cursor.execute(count_notifs)
    row = cursor.fetchone()
    num_notifs = int(row[0])

    return num_notifs


def main():
    logging.basicConfig(
        filename="orbital_prediction_service.log",
        format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        level=logging.DEBUG,
        filemode="a",
    )

    stream_handler = logging.StreamHandler()
    stream_handler.setFormatter(
        logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
    )

    logger.addHandler(logging.StreamHandler())
    logger.info("Parsing Environment Variables")
    load_dotenv()
    period = os.getenv("PASS_PERIOD")

    if type(period) is str:
        period = int(period)
    else:
        print("Failed to parse PERIOD env var")
        return

    while True:
        logger.info("Opening connection to database")
        conn = sqlite3.connect("./vertex.db")
        cursor = conn.cursor()
        calc_period = dt.datetime.now(dt.UTC) + dt.timedelta(seconds=period)
        logger.info(f"Next pass computation approximately: {calc_period} UTC")
        logger.info("Querying Notifications")
        num_notifs = CountNotifs(cursor)
        logger.info(f"Current Notifications: {num_notifs}")

        if num_notifs > 0:
            logger.info("Querying Satellites and Stations")
            sats = QuerySats(cursor)
            stns = QueryStns(cursor)
            logger.info(f"Received {len(sats)} Satellites and {len(stns)} Stations")
            logger.info("Finding max horizon for next prediction")
            max_horizon = SelectMaxHorizon(cursor)
            logger.info(f"Max Horizon: {max_horizon} hrs")
            logger.info("Deleting Old Passes")
            DeletePasses(conn, cursor)
            logger.info("Deleted Old Passes")
            logger.info("Computing New Passes")
            passes = ComputePasses(stns, sats, max_horizon)
            logger.info("Inserting Newly Computed Passes Into Database")
            InsertPasses(conn, cursor, passes)
            logger.info(f"Inserted {len(passes)} Passes into Database")

        logger.info("Deleting Notifications")
        DeleteOrbitalNotifs(conn, cursor)
        logger.info("Deleted Notifications")
        logger.info("Deleting Horizon")
        DeleteHorizons(conn, cursor)
        logger.info("Deleted Horizon")
        logger.info("Closing Database connection")
        conn.close()

        now = dt.datetime.now(dt.UTC)
        if now < calc_period:
            diff = calc_period - now
            logger.info(f"Sleeping for {diff.seconds}")
            time.sleep(diff.seconds)


if __name__ == "__main__":
    main()
