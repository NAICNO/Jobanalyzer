# NAIC Jobanalyzer Dashboard

This is the second iteration of the dashboard, which is a complete rewrite of the first iteration.
This will display the data produced by naicreport and will be used to query the data interactively.

## Prerequisites

Before you begin, ensure you have met the following requirements:
- Node.js (preferably the latest LTS version). You can download it from [Node.js official site](https://nodejs.org/).
- A package manager like Yarn (preferred) or npm. 

## Setting up your local environment

To set up your local environment, follow these steps:

1. Clone the repository to your local machine.
2. Navigate to the project directory.
```commandline
cd code/dashboard-2
```
3. Install the dependencies.
```commandline
yarn install
```

## Running the application locally

To run the application locally, follow these steps:

1. Start the development server.
```commandline
yarn dev
```
2. Open your browser and navigate to `http://localhost:5173`. Default port is 5173. If port 5173 is not available, the application will automatically switch to the next available port. You can see the port number in the terminal where you started the development server.

## Building the application for production

To build the application for production, follow these steps:

```commandline
yarn build
```

This will create a `build` directory in the project root. You can deploy the contents of this directory to a web server.
