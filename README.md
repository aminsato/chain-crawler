# Chain Crawler

Chain Crawler is a project that allows you to crawl information about various chains and store it in a LevelDB database. Additionally, it provides a REST API that enables users to retrieve results from the database.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Installation](#installation)
## Introduction

Chain Crawler is designed to crawl and store information about various chains, providing a centralized repository for chain-related data.

## Features

- **Chain Information Crawling:** Retrieve detailed information about various chains.
- **LevelDB Database:** Store and manage the crawled data in a LevelDB database.
- **Docker Compose Support:** Easily deploy and run the crawler using Docker Compose.
- **REST API:** Access the crawled data through a RESTful API.
  
## Getting Started


### Installation

1. Clone the repository.
2. Retrieve an API token and node address from [Infura](https://app.infura.io/).
3. Place the token and node address into `/crw/main.go`


var defaultNodeAddress = "https://mainnet.infura.io/v3/YOUR_INFURA_API_TOKEN"


