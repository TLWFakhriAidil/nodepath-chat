// Simple script to test database connection
import { execSync } from 'child_process';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import dotenv from 'dotenv';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

console.log('Testing database connection...');

// Load environment variables from .env file if it exists
if (fs.existsSync(path.join(__dirname, '.env'))) {
  console.log('Loading environment variables from .env file');
  dotenv.config();
}

// Display current environment variables
console.log('Environment variables:');
console.log(`DB_HOST: ${process.env.DB_HOST || 'not set'}`);
console.log(`DB_NAME: ${process.env.DB_NAME || 'not set'}`);
console.log(`DB_USER: ${process.env.DB_USER || 'not set'}`);
console.log(`DB_PASSWORD: ${process.env.DB_PASSWORD ? '[masked]' : 'not set'}`);
console.log(`DB_PORT: ${process.env.DB_PORT || 'not set'}`);
console.log(`VITE_DB_HOST: ${process.env.VITE_DB_HOST || 'not set'}`);
console.log(`VITE_DB_NAME: ${process.env.VITE_DB_NAME || 'not set'}`);
console.log(`VITE_DB_USER: ${process.env.VITE_DB_USER || 'not set'}`);
console.log(`VITE_DB_PASSWORD: ${process.env.VITE_DB_PASSWORD ? '[masked]' : 'not set'}`);
console.log(`VITE_DB_PORT: ${process.env.VITE_DB_PORT || 'not set'}`);

// Check if MySQL environment variables are set
const mysqlConfigured = process.env.VITE_DB_HOST || process.env.DB_HOST;

if (!mysqlConfigured) {
  console.log('⚠️ WARNING: MySQL environment variables are not set!');
  console.log('Please set the following environment variables in your .env file:');
  console.log('  VITE_DB_HOST (for local development)');
  console.log('  VITE_DB_NAME (for local development)');
  console.log('  VITE_DB_USER (for local development)');
  console.log('  VITE_DB_PASSWORD (for local development)');
  console.log('  VITE_DB_PORT (for local development)');
  console.log('OR');
  console.log('  DB_HOST (for Railway deployment)');
  console.log('  DB_NAME (for Railway deployment)');
  console.log('  DB_USER (for Railway deployment)');
  console.log('  DB_PASSWORD (for Railway deployment)');
  console.log('  DB_PORT (for Railway deployment)');
  console.log('\nSee .env.example for reference.');
} else {
  console.log('✅ MySQL environment variables are set.');
}

// Test PHP database connection
console.log('\nTesting PHP database connection...');
try {
  const phpOutput = execSync('php db-test.php').toString();
  console.log('PHP database test result:');
  console.log(phpOutput);
} catch (error) {
  if (error.stderr && error.stderr.toString().includes('not recognized')) {
    console.log('❌ PHP is not installed or not in your PATH.');
    console.log('Please install PHP to test the database connection with db-test.php.');
    console.log('You can download PHP from: https://windows.php.net/download/');
  } else {
    console.error('Error running PHP database test:', error.message);
    if (error.stdout) {
      console.log('Output:', error.stdout.toString());
    }
    if (error.stderr) {
      console.log('Error output:', error.stderr.toString());
    }
  }
}

console.log('\nDatabase connection test complete.');