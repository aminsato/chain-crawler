# Chain Crawler

Chain Crawler is a project that allows you to crawl information about various chains and store it in a LevelDB database. Additionally, it provides a REST API that enables users to retrieve results from the database.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Installation](#installation)
- [Results](#Results)

## Introduction

Chain Crawler is designed to crawl and store information about various chains, providing a centralized repository for chain-related data.

## Features

- **Chain Information Crawling:** Retrieve detailed information about various chains.
- **LevelDB Database:** Store and manage the crawled data in a LevelDB database.
- **Docker Compose Support:** Easily deploy and run the crawler using Docker Compose.
- **REST API:** Access the crawled data through a RESTful API.
- **gRPC Integration:** Seamlessly access the crawled data through a gRPC-based interface.
- **Cobra Command Line Interface (CLI):** Execute the program effortlessly using intuitive commands and flags facilitated by the Cobra library.
- **API Documentation in YAML:** Provide comprehensive and structured documentation for the API using YAML format.

## Getting Started

### Installation

1. Clone the repository.
2. Retrieve an API token and node address from [Infura](https://app.infura.io/).
3. Place the token and node address into `/crw/main.go`


var defaultNodeAddress = "https://mainnet.infura.io/v3/YOUR_INFURA_API_TOKEN"


## Results

Here are some of my interesting summary results about the contracts in 2023.

The top accounts that paid the highest fees in 2023.


| account_address                            | total_paid_fee  |
|--------------------------------------------|-----------------|
| 0xae2fc483527b8ef99eb5d9b44875f005ba1fae13 | 52117.202360228 |
| 0xc1b634853cb333d3ad8663715b08f41a3aec47cc | 24556.350469831 |
| 0x3527439923a63f8c13cf72b8fe80a77f6e572092 | 14117.65930822  |
| 0x6887246668a3b87f54deb3b94ba47a6f63f32985 | 13005.099261765 |
| 0x16d5783a96ab20c9157d7933ac236646b29589a4 | 12907.201414577 |


The top contract accounts that had been spent the highest fees in 2023.


| contract_address                               | total_spend_fee  |        
|------------------------------------------------|------------------|
| 0x7a250d5630b4cf539739df2c5dacb4c659f2488d     | 105625.339320266 |  
| 0xef1c6e67703c7bd7107eed8303fbe6ec2554bf6b     | 87081.423163584  |  
| 0x3fc91a3afd70395cd496c647d5a6cc9d4b2b7fad     | 68250.854763935  |  
| 0xdac17f958d2ee523a2206206994597c13d831ec7     | 53670.835072488  |  
| 0x6b75d8af000000e20b7a7ddf000ba900b4009a80     | 52473.171381869  |


