-- Drop indexes first
DROP INDEX IF EXISTS idx_objects_name;
DROP INDEX IF EXISTS idx_chunks_hash;
DROP INDEX IF EXISTS idx_chunks_object_id;
DROP INDEX IF EXISTS idx_chunk_node_chunk_id;
DROP INDEX IF EXISTS idx_chunk_node_node_id;

-- Drop tables in correct order to avoid foreign key constraint issues
DROP TABLE IF EXISTS chunk_node_locations;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS masters;
DROP TABLE IF EXISTS chunks;
DROP TABLE IF EXISTS objects;
