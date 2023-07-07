CREATE DATABASE lru_mapper;
CREATE DATABASE lru_redirect;
CREATE DATABASE lru_analytic;

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Enable pg_stat_statements in the lru_mapper database
\c lru_mapper

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Enable pg_stat_statements in the lru_redirect database
\c lru_redirect

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Enable pg_stat_statements in the lru_analytic database
\c lru_analytic

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;