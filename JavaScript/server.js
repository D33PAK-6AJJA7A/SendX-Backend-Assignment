const express = require('express');
const fs = require('fs');
const path = require('path');
const http = require('https');

const app = express();
const workerPool = [];
const cache = {};
const MAX_WORKERS = 10; // maximum number of active workers

// creates cache folder if it doesnot exists to store downloaded files
const cacheFolder = 'cache';
if (!fs.existsSync(cacheFolder)) {
    fs.mkdirSync(cacheFolder);
    console.log(`Created ${cacheFolder} folder`);
} else {
    console.log(`${cacheFolder} folder already exists`);
}

app.get('/download', (req, res) => {
    // parse the query string to get the URL and retry limit
    const webpageURL = req.query.url;
    const retryLimit = parseInt(req.query.retryLimit);

    // make sure retry limit does not exceed 10
    if (retryLimit > 10) {
        retryLimit = 10;
    }

    // check if the webpage has been requested in the last 24 hours
    const cachedPage = cache[webpageURL];
    if (cachedPage && Date.now() - cachedPage.timestamp < 24 * 60 * 60 * 1000) {
        // update date of cached web page as it is accessed again
        cache[webpageURL] = {
            timestamp: Date.now(),
            html: cachedPage.html,
        };
        // serve the webpage from the cache
        console.log(`Web page already exists`);
        res.send(cachedPage.html);
        return;
    }

    // download the webpage
    let retries = 0;
    const worker = () => {
        http.get(webpageURL, (response) => {
            let data = '';
            response.on('data', (chunk) => {
                data += chunk;
            });
            response.on('end', () => {
                // cache the webpage
                cache[webpageURL] = {
                    timestamp: Date.now(),
                    html: data,
                };

                // download the webpage to the local file system
                const nameStr = "cache\\" + webpageURL.toString().replace(':', '^').replace(/\//g, '_');
                const filePath = path.join(__dirname, `${nameStr}.html`);
                fs.writeFile(filePath, data, (err) => {
                    if (err) {
                        console.error(err);
                        res.status(500).send('Error downloading webpage');
                    } else {
                        console.log(`cached webpage successfully`)
                        res.send(data);
                    }
                });
            });
        }).on('error', (err) => {
            console.error(err);
            if (retries < retryLimit) {
                retries++;
                worker();
            } else {
                res.status(500).send('Error downloading webpage');
            }
        });
    };

    // add the worker to the pool and start it
    workerPool.push(worker);
    if (workerPool.length > MAX_WORKERS) {
        // if the number of active workers exceeds the maximum, remove the oldest worker from the pool
        workerPool.shift()();
    } else {
        worker();
    }
});

app.listen(3000, () => {
    console.log('Server listening on port 3000');
});