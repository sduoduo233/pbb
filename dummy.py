import sqlite3
import time
import random
import math

conn = sqlite3.connect("sqlite.db")
cur = conn.cursor()

cur.execute("DELETE FROM service_metrics")

now = int(time.time()) // 300 * 300
for t in range(now - 3 * 24 * 3600, now, 300):
    l = sorted([random.randint(60000, 62000) for _ in range(20)])

    if random.random() < 0.1:
        l.append(110000)

    y = math.sin(t / 3600)
    minn = int(y * 200 + 1000)
    maxx = int(y * 200 + 1500)
    median = int(minn / 2 + maxx / 2)
    loss = (t % (3600 * 4)) / (3600 * 4)

    # if t // 3600 & 1:
    # continue

    cur.execute(
        f"INSERT INTO service_metrics (created_at, timestamp, `from`, `to`, min, max, loss, avg, median) VALUES ({t}, {t}, {1}, {1}, {minn}, {maxx}, {loss}, {int(sum(l)/20)}, {median})"
    )

conn.commit()
