/*
 *   Copyright 2020 Tero Vierimaa
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package migrations

// modify metadata primary key, add bookmark
// rebuild triggers

const v5 = `
CREATE TABLE metadata_copy 
( 
	bookmark    INT,
    key         TEXT NOT NULL,
    key_lower   TEXT NOT NULL,
    value       TEXT NOT NULL,
    value_lower TEXT NOT NULL,
	CONSTRAINT metadata_pk
        PRIMARY key (bookmark, key_lower)
);
INSERT INTO metadata_copy(bookmark, key, key_lower, value, value_lower)
SELECT 
	bookmark, key, key_lower, value, value_lower 
FROM metadata
ORDER BY bookmark;
DROP TABLE metadata;
ALTER TABLE metadata_copy RENAME TO metadata;

CREATE TRIGGER create_metadata_fts
    AFTER INSERT ON metadata BEGIN
    INSERT INTO metadata_fts(id, key, value)
    VALUES (new.bookmark, new.key, new.value);
END;
CREATE TRIGGER update_metadata_fts
    AFTER UPDATE ON metadata BEGIN
    UPDATE metadata_fts SET
      id = new.bookmark,
        key = new.key,
        value = new.value
    WHERE id = old.bookmark AND key = old.key;
END;
CREATE TRIGGER delete_metadata_fts
    AFTER DELETE ON metadata BEGIN
    DELETE FROM metadata_fts
    WHERE id = old.bookmark;
END;
`
