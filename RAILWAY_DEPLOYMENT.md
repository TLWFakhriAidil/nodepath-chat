# Railway Deployment Guide

## Environment Variables Setup

To properly deploy this application on Railway, you need to set up the following environment variables in your Railway project:

### Required Environment Variables

| Variable | Description | Example Value |
|----------|-------------|---------------|
| `DB_HOST` | MySQL database host | 159.89.198.71 |
| `DB_NAME` | MySQL database name | admin_railway |
| `DB_USER` | MySQL database user | admin_aqil |
| `DB_PASSWORD` | MySQL database password | admin_aqil |
| `DB_PORT` | MySQL database port | 3306 |

## Setting Environment Variables in Railway

1. Go to your Railway project dashboard
2. Click on the "Variables" tab
3. Add each of the environment variables listed above
4. Click "Deploy" to apply the changes

## Troubleshooting

### 502 Bad Gateway Error

If you encounter a 502 Bad Gateway error, it could be due to one of the following reasons:

1. **Missing Environment Variables**: Ensure all required environment variables are set correctly.
2. **Database Connection Issues**: Verify that the database credentials are correct and that the database is accessible from Railway.
3. **Application Not Listening on Correct Port**: The application must listen on the port specified by Railway (4173 according to your configuration).

### Database Connection Issues

To troubleshoot database connection issues:

1. Check the Railway logs for any error messages related to database connections.
2. Verify that the database credentials are correct.
3. Ensure that the database server allows connections from Railway's IP addresses.
4. Try running the `db-test.php` endpoint to test the database connection.

## Logs and Debugging

To view logs and debug your application:

1. Go to your Railway project dashboard
2. Click on the "Logs" tab
3. Look for any error messages related to database connections or environment variables

## Local Testing

Before deploying to Railway, you can test your application locally:

1. Create a `.env` file based on the `.env.example` template
2. Run `npm run test:db` to test the database connection
3. Run `npm run dev` to start the development server

## Deployment Process

Railway uses the following files for deployment:

- `railway.toml`: Configuration for Railway deployment
- `Dockerfile`: Instructions for building the Docker container
- `start.sh`: Script that runs when the container starts

The deployment process is as follows:

1. Railway builds the Docker container using the Dockerfile
2. The container starts and runs the `start.sh` script
3. The script checks for environment variables and sets defaults if needed
4. The application starts and attempts to connect to the database

## Additional Resources

- [Railway Documentation](https://docs.railway.app/)
- [MySQL Connection Troubleshooting](https://dev.mysql.com/doc/refman/8.0/en/problems-connecting.html)