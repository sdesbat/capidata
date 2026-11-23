import os
from sqlalchemy import create_engine, text

URL = os.getenv("DATABASE_URL", "postgresql+psycopg://badger:badger@localhost:5432/badger")
engine = create_engine(URL, future=True)

with engine.begin() as connection:
    connection.execute(text("CREATE TABLE IF NOT EXISTS audit_demo (id integer primary key, tenant text not null, status text not null, amount integer not null)"))
    connection.execute(text("INSERT INTO audit_demo (id, tenant, status, amount) VALUES (:id, :tenant, :status, :amount) ON CONFLICT (id) DO UPDATE SET status = EXCLUDED.status, amount = EXCLUDED.amount"), {"id": 1, "tenant": "demo", "status": "pending", "amount": 10})

with engine.begin() as connection:
    connection.execute(text("UPDATE audit_demo SET status = :status, amount = :amount WHERE id = :id"), {"status": "committed", "amount": 20, "id": 1})

print("INSERT and UPDATE committed")
