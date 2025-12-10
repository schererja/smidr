export class BuildQueue {
    constructor(maxConcurrent, logger) {
        this.waiting = [];
        this.semaphore = maxConcurrent;
        this.logger = logger;
    }
    async acquire() {
        if (this.semaphore > 0) {
            this.semaphore--;
            return;
        }
        return new Promise((resolve) => {
            this.waiting.push(() => {
                this.semaphore--;
                resolve();
            });
        });
    }
    release() {
        if (this.waiting.length > 0) {
            const next = this.waiting.shift();
            if (next) {
                next();
            }
        }
        else {
            this.semaphore++;
        }
    }
}
