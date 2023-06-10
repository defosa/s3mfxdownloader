S3MXFDownloader is a powerful tool for efficiently downloading MXF (Material Exchange Format) files from an Amazon S3 bucket. This program ensures that only MXF files are retrieved while also logging their names to a PostgreSQL database table for easy tracking and management.


## Features

  -  Seamless integration with Amazon S3: Connects to the specified S3 bucket using the AWS SDK to search for and download MXF files.
  -  Automatic file filtering: Scans the bucket for MXF files based on a specified prefix and only downloads those files, ensuring efficient file retrieval.
  -  Concurrent downloads: Optimizes the download process by supporting concurrent file downloads with customizable concurrency levels.
  -  Duplicate file prevention: Checks filenames against a PostgreSQL database table before downloading to avoid duplicate files, enhancing file management efficiency.
  -  Database logging: Establishes a connection to a PostgreSQL database and logs the names of downloaded MXF files to a designated table for easy tracking and retrieval.

## Prerequisites

- AWS credentials: Ensure you have valid AWS access key and secret key with appropriate permissions to access the target S3 bucket.
- PostgreSQL database: Set up a PostgreSQL database and provide the necessary connection string details in the program's environment variables.

## Usage

1. Install the program's dependencies by running `go mod tidy` in the project directory.
2. Set the required environment variables:
   - `YOUR_AWS_ACCESS_KEY`: Your AWS access key.
   - `YOUR_AWS_SECRET_KEY`: Your AWS secret key.
   - `YOUR_BUCKET_NAME`: The name of the target S3 bucket.
   - `S3ENDURL`: The endpoint URL for the S3 service.
   - `REGION`: The AWS region for the S3 bucket.
   - `PREFIX`: The prefix to filter MXF files in the S3 bucket.
   - `DB_CONNECTION_STRING`: The connection string for the PostgreSQL database.
3. Run the program using `go run main.go`.
4. Send a POST request to the program's endpoint to initiate the download process, for example:   
```

    curl -i --header 'Content-Type: application/json' --request POST --data '{"START"}' http://localhost:8080
```




### Example of creating a table in a database
```

CREATE TABLE IF NOT EXISTS transcode_jobs (
			ID VARCHAR(255) PRIMARY KEY,
		);

```
