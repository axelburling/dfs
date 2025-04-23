CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create Objects Table
CREATE TABLE objects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    total_size BIGINT NOT NULL,
    total_chunks INTEGER NOT NULL,
    content_type VARCHAR(255) NOT NULL
);

-- Create Chunks Table
CREATE TABLE chunks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    object_id UUID NOT NULL,
    chunk_index INTEGER NOT NULL,
    chunk_size BIGINT NOT NULL,
    chunk_hash VARCHAR(64) NOT NULL,
    CONSTRAINT fk_object FOREIGN KEY (object_id) REFERENCES objects(id) ON DELETE CASCADE
);

-- Create Masters Table
CREATE TABLE masters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(255) NOT NULL,
    grpc_address VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_healthy BOOLEAN DEFAULT true
);

-- Create Nodes Table
CREATE TABLE nodes (
    id UUID PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    grpc_address VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_healthy BOOLEAN DEFAULT true,
    total_space BIGINT NOT NULL,
    free_space BIGINT NOT NULL,
    readonly BOOLEAN DEFAULT false,
    controller_master_id UUID DEFAULT NULL REFERENCES masters(id) ON DELETE SET NULL
);

-- Create Mapping Table Between Chunks & Nodes
CREATE TABLE chunk_node_locations (
    chunk_id UUID NOT NULL,
    node_id UUID NOT NULL,
    PRIMARY KEY (chunk_id, node_id),
    CONSTRAINT fk_chunk FOREIGN KEY (chunk_id) REFERENCES chunks(id) ON DELETE CASCADE,
    CONSTRAINT fk_node FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
);

-- Indexes for Performance
CREATE INDEX idx_objects_name ON objects(name);
CREATE INDEX idx_chunks_hash ON chunks(chunk_hash);
CREATE INDEX idx_chunks_object_id ON chunks(object_id);
CREATE INDEX idx_chunk_node_chunk_id ON chunk_node_locations (chunk_id);
CREATE INDEX idx_chunk_node_node_id ON chunk_node_locations (node_id);
