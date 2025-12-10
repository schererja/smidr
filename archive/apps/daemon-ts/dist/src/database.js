import initSqlJs from "sql.js";
import { promises as fs } from "fs";
export class DB {
    constructor(dbPath = "smidr.db") {
        this.db = null;
        this.isInitialized = false;
        this.dbPath = dbPath;
    }
    async initialize() {
        if (this.isInitialized)
            return;
        const SQL = await initSqlJs();
        try {
            const fileBuffer = await fs.readFile(this.dbPath);
            this.db = new SQL.Database(fileBuffer);
        }
        catch {
            this.db = new SQL.Database();
        }
        const schema = `
      CREATE TABLE IF NOT EXISTS builds (
        id TEXT PRIMARY KEY,
        customer_id TEXT NOT NULL,
        target TEXT NOT NULL,
        config_path TEXT NOT NULL,
        state TEXT NOT NULL,
        start_time INTEGER NOT NULL,
        end_time INTEGER,
        exit_code INTEGER
      );

      CREATE TABLE IF NOT EXISTS build_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        build_id TEXT NOT NULL,
        timestamp INTEGER NOT NULL,
        stream TEXT NOT NULL,
        message TEXT NOT NULL,
        FOREIGN KEY (build_id) REFERENCES builds(id)
      );

      CREATE INDEX IF NOT EXISTS idx_build_logs_build_id ON build_logs(build_id);
      CREATE INDEX IF NOT EXISTS idx_builds_customer ON builds(customer_id);
    `;
        this.db.run(schema);
        this.isInitialized = true;
    }
    execute(sql, params = []) {
        if (!this.db)
            throw new Error("Database not initialized");
        this.db.run(sql, params);
    }
    query(sql, params = []) {
        if (!this.db)
            throw new Error("Database not initialized");
        const stmt = this.db.prepare(sql);
        stmt.bind(params);
        const results = [];
        while (stmt.step()) {
            results.push(stmt.getAsObject());
        }
        stmt.free();
        return results;
    }
    saveBuild(id, target, state, startTime, customerId, configPath) {
        this.execute(`
      INSERT OR REPLACE INTO builds (id, customer_id, target, config_path, state, start_time)
      VALUES (?, ?, ?, ?, ?, ?)
    `, [id, customerId || "", configPath || "", target, state, startTime]);
        this.saveToFile();
    }
    getBuild(buildId) {
        const results = this.query("SELECT * FROM builds WHERE id = ?", [buildId]);
        return results.length > 0 ? results[0] : null;
    }
    listBuilds(customerId, limit = 10) {
        const results = this.query(`
      SELECT * FROM builds WHERE customer_id = ? ORDER BY start_time DESC LIMIT ?
    `, [customerId, limit]);
        return results;
    }
    updateBuildState(buildId, state, endTime, exitCode) {
        this.execute(`
      UPDATE builds SET state = ?, end_time = ?, exit_code = ? WHERE id = ?
    `, [state, endTime || null, exitCode || null, buildId]);
        this.saveToFile();
    }
    saveLogs(buildId, logs) {
        for (const log of logs) {
            this.execute(`
        INSERT INTO build_logs (build_id, timestamp, stream, message)
        VALUES (?, ?, ?, ?)
      `, [buildId, log.timestamp, log.stream, log.message]);
        }
        this.saveToFile();
    }
    getLogs(buildId) {
        const results = this.query(`
      SELECT timestamp, stream, message FROM build_logs WHERE build_id = ? ORDER BY timestamp ASC
    `, [buildId]);
        return results;
    }
    async saveToFile() {
        if (!this.db)
            return;
        try {
            const data = this.db.export();
            const buffer = Buffer.from(data);
            await fs.writeFile(this.dbPath, buffer);
        }
        catch (err) {
            console.error("Failed to save database:", err);
        }
    }
    async close() {
        await this.saveToFile();
        this.db?.close();
    }
}
