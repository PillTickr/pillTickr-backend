CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE medicines (
    medicine_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    dosage VARCHAR(50),         -- e.g. "1 pill", "5ml"
    instructions TEXT,          -- e.g. "After meals"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);


CREATE TABLE schedules (
    schedule_id INTEGER PRIMARY KEY AUTOINCREMENT,
    medicine_id INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,                       -- NULL = ongoing
    frequency TEXT NOT NULL CHECK (frequency IN ('daily','weekly','custom')),
    times_per_day INTEGER NOT NULL,          -- e.g. 3 times a day
    FOREIGN KEY (medicine_id) REFERENCES medicines(medicine_id) ON DELETE CASCADE
);


CREATE TABLE schedule_times (
    time_id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER NOT NULL,
    intake_time TIME NOT NULL,           -- e.g. "08:00", "14:00"
    FOREIGN KEY (schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE
);


CREATE TABLE reminders (
    reminder_id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER NOT NULL,
    reminder_datetime DATETIME NOT NULL,
    status TEXT CHECK (status IN ('taken', 'pending', 'missed')) DEFAULT 'pending',
    taken_at DATETIME,
    FOREIGN KEY (schedule_id) REFERENCES schedules(schedule_id) ON DELETE CASCADE
);

