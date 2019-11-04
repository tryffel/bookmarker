/*
 *
 *  Copyright 2019 Tero Vierimaa
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *
 */

package migrations

const v1 = `
CREATE TABLE bookmarks (
	id INTEGER
		CONSTRAINT bookmarks_pk
			PRIMARY KEY autoincrement,
	name STRING NOT NULL,
	lower_name STRING NOT NULL,
	description STRING,
	content STRING NOT NULL,
	project STRING,
	created_at TIMESTAMP DEFAULT CURRENT_DATE,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE "tags" (
	name STRING NOT NULL,
	id INTEGER
		CONSTRAINT tags_pk
			PRIMARY KEY AUTOINCREMENT
);


CREATE TABLE bookmark_tags (
	bookmark INTEGER CONSTRAINT bookmark
			REFERENCES bookmarks,
	tag INTEGER CONSTRAINT bookmark_tags_pk
			PRIMARY key
		CONSTRAINT tags
			REFERENCES tags
);
`
