import sqlite3
import time
import random

conn = sqlite3.connect("sqlite.db")
cur = conn.cursor()

cur.execute("DELETE FROM service_metrics")

now = int(time.time()) // 300 * 300
for t in range(now - 3 * 24 * 3600, now,  300):
    l = sorted([random.randint(500, 5000) for _ in range(20)])
    
    cur.execute(f"INSERT INTO service_metrics (created_at, timestamp, `from`, `to`, min, max, loss, avg, median) VALUES ({t}, {t}, {1}, {6}, {min(l)}, {max(l)}, 0, {int(sum(l)/20)}, {l[10]})")

conn.commit()