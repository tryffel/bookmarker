/*
 *   Copyright 2019 Tero Vierimaa
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

const v4 = `
CREATE VIRTUAL TABLE 
bookmark_fts
USING fts5(
	id UNINDEXED,
	name,
	description,
	content,
	project);

CREATE VIRTUAL TABLE 
metadata_fts 
USING fts5(
	id UNINDEXED, 
	key, 
	value);

-- index existing data to fts tables
INSERT INTO bookmark_fts(id, name, description,content,project)
SELECT
	id, name, description, content, project
FROM bookmarks;

INSERT INTO metadata_fts(id, key, value)
SELECT
	bookmark, key, value 
FROM metadata;



-- triggers to keep fts updated
CREATE TRIGGER create_bookmark_fts
    AFTER INSERT ON bookmarks BEGIN
    INSERT INTO bookmark_fts(id, name, description, content, project)
        VALUES (new.id, new.name, new.description, new.content, new.project);
END;

CREATE TRIGGER update_bookmark_fts
    AFTER UPDATE ON bookmarks BEGIN
    UPDATE bookmark_fts SET
                            name = new.name,
                            description = new.description,
                            content = new.content,
                            project = new.project
        WHERE id = new.id;
END;

CREATE TRIGGER delete_bookmark_fts
    AFTER DELETE ON bookmarks BEGIN
        DELETE FROM bookmark_fts
        WHERE id = old.id;
END;

-- fill bookmark_fts with current bookmarks
INSERT INTO bookmark_fts(id, name, description,content,project)
SELECT
    id, name, description, content, project
FROM bookmarks;

-- fill description_lower row
UPDATE bookmarks SET
description_lower = LOWER(description) WHERE true;
`
