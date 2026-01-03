# Scalable E-Commerce Application on AWS (Online Boutique)
## Group Members
- Muhammad Hamza (BSSE23001)
- Muhammad Sarmad (BSSE23013)
- Hafiz Azfar Umar Sheikh (BSSE23064)

## Project Overview
This project implements the "Online Boutique" - a cloud-native, microservices-based e-commerce platform designed for high availability and fault tolerance. We migrated the traditional monolithic architecture into a decoupled ecosystem on Amazon Web Services (AWS) to address scalability issues and single points of failure.
Our implementation focuses on a "Serverless-First" approach using Amazon ECS with Fargate, eliminating the need for manual server management while ensuring 99.9% availability through multi-AZ deployment in us-east-1.

## Infrastructure Highlights
- Compute: 11 Microservices running as serverless tasks on AWS Fargate.
- Networking:
  - Public ALB (Layer 7): Handles incoming user traffic and routes it to the Frontend.
  - Internal NLB (Layer 4): Manages ultra-low latency, secure TCP communication between internal services (Frontend → Backend).
- Database & Storage:
  - Amazon DynamoDB: NoSQL database for the Product Catalog (Partition Key: ProductId).
  - Amazon ElastiCache (Redis): In-memory data store for managing high-speed user sessions and cart data.
- Security:
  - IAM Roles: Least Privilege access using custom LabRole for ECS task execution.
  - Security Groups: Micro-segmentation ensuring the backend only accepts traffic from the Internal NLB.

## Tech Stack
- Cloud Provider: AWS (ECS, Fargate, VPC, DynamoDB, ElastiCache, ECR, CloudWatch)
- Containerization: Docker
- Orchestration: Amazon ECS
- Languages: Go, Java, Python, c#, Node.js