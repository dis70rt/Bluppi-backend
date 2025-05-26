# Discontinued Cassandra Implementation

## Overview
This directory contains a discontinued implementation of a chat database system using **Apache Cassandra**. The implementation was designed to provide data storage and retrieval functionality for a chat application, including features for:

- Conversation management  
- Participant tracking  
- Message handling

## Reason for Discontinuation

The implementation has been discontinued for the following reasons:

- **Resource Requirements**:  
  Cassandra's high memory requirements exceeded the available resources on our target deployment environment.

- **Query Optimization Challenges**:  
  The implementation relied heavily on Cassandra's `ALLOW FILTERING` clause, which is not recommended for production use due to performance implications. As noted in the code comments, proper partitioning strategies would be required for an efficient implementation.

- **Technical Expertise Gap**:  
  Our team lacked the specialized knowledge required to optimize Cassandra for this specific use case, particularly regarding efficient filtering and query design patterns.

## Implementation Features

The implementation included functionality for:

- Creating and managing conversations (group and individual)
- Managing conversation participants
- Storing and retrieving messages
- Message status tracking
- Pagination support for message history

## Alternatives

We are exploring alternative database solutions that better fit our resource constraints and technical expertise.  
If you're interested in this project, please refer to the current implementation in the **parent directory**.
