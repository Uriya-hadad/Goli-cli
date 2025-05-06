// create simple server
const env = require('@sap/xsenv');
const express = require('express');
const path = require('path');
const {Pool} = require('pg');
const {createMacPackage, createWinPackage} = require('./artifactory');

const app = express();
let pool;


function getDbCredentials() {
    credentials = env.cfServiceCredentials({label: 'postgresql-db'});

    return credentials;
}
async function initTable() {
    query = `CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        mail VARCHAR(255) NOT NULL,
        auth boolean DEFAULT false,
        role VARCHAR(255) DEFAULT 'developer'
    )`;
    await pool.query(query);
}

async function addUser(mail) {
    query = `INSERT INTO users (mail) VALUES ('${mail}')`
    try {
        await pool.query(query);
    } catch (err) {
        console.log('addUser - failed to insert user:', err);
        throw new Error('failed to insert user');
    }
}

async function getUser(mail) {
    query = `SELECT * FROM users WHERE mail = '${ mail }'`;
    console.log(query)
    res = await pool.query(query)
    return res.rows[0];
}

async function authUser(mail,role) {
    query = `UPDATE users SET role = '${role}', auth = true WHERE mail = '${mail}'`
    await pool.query(query, (err, res) => {
        if (err) {
            console.log('authUser - failed to update user:', err);
            throw new Error('failed to update user');
        }
    });
    
}

app.get('/goli/auth', async (req, res) => {
    const mail = req.query.mail;
    const role = req.query.role;
    try {
        await authUser(mail,role);
    } catch (err) {
        console.log('Failed to authenticate user:', err);
        return res.send('Failed to authenticate user');
    }
    res.send('User authenticated');
});

(async function init() {

    // get db credentials
    let credentials = getDbCredentials();
    if (!credentials) {
        logger.error(null, 'init - failed to initialize DB credentials');
        throw new Error('failed to initialize DB credentials');
    }

    // config db schema
    let connOptions = {
        user: credentials.username,
        host: credentials.hostname,
        database: credentials.dbname,
        password: credentials.password,
        port: credentials.port
    };

    if (credentials.sslcert && credentials.sslrootcert) {
        connOptions.ssl = {
            ca: credentials.sslrootcert,
            cert: credentials.sslcert,
            key: credentials.client_key,
            checkServerIdentity: () => {
            }
        };
    } else if (credentials.ca_base64) {
        connOptions.ssl = {
            ca: credentials.ca_base64,
            checkServerIdentity: () => {
            }
        };
    }

    // start migration
    try {
        pool = new Pool(connOptions);
        await initTable(pool);
        console.log('init - successfully run db migration files');
    } catch (err) {
        console.log(null, 'init - failed to run DB migration files:', err);
        throw new Error('failed to run migration files');
    }
})();

app.get('/goli/register', async (req, res) => {
    const mail = req.query.mail;
    if (!mail) {
        return res.send('Mail is required');
    }
    try {
       await addUser(mail);
    } catch (err) {
        console.log('Failed to add user:', err);
        return res.send('Failed to add user');
    }
    res.send('User registered');
})
app.get('/goli/user', async (req, res) => {
    const mail = req.query.mail;
    try {
        user = await getUser(mail);
    } catch (err) {
        console.log('Failed to get user:', err);
        return res.send('Failed to get user');
    }
    if (user && user.auth !== undefined) {
        user.role = Buffer.from(user.role).toString('base64')
        return res.json(user);
        // return res.send(user.auth);
    }
    res.send('User not found');
})

app.get('/goli/mac', async (req, res) => {
    createMacPackage(res)
});

app.get('/goli/win', async (req, res) => {
    createWinPackage(res)
});

app.get('/goli/goli-completion-win', (req, res) => {
    res.download(path.join(__dirname, '../completionFiles/powershell'), 'goli-completion-win');
});

app.get('/goli/goli-completion-mac', (req, res) => {
    res.download(path.join(__dirname, '../completionFiles/zsh'), 'goli-completion-mac');
});


app.use('/goli',express.static(path.join(__dirname, '../public')));

const port = process.env.PORT || 3000;

app.listen(port, () => {
    console.log('Server started');
})

