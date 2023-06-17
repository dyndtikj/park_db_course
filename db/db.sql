CREATE EXTENSION IF NOT EXISTS citext;

CREATE UNLOGGED TABLE IF NOT EXISTS "user"
(
    id       bigserial                  NOT NULL PRIMARY KEY,
    nickname citext collate "ucs_basic" NOT NULL UNIQUE,
    fullname citext                     NOT NULL,
    about    text,
    email    citext                     NOT NULL UNIQUE
);

CREATE UNLOGGED TABLE IF NOT EXISTS forum
(
    id      bigserial NOT NULL PRIMARY KEY,
    title   text      NOT NULL,
    "user"  citext    NOT NULL,
    slug    citext    NOT NULL UNIQUE,
    posts   bigint DEFAULT 0,
    threads int    DEFAULT 0
);

CREATE UNLOGGED TABLE IF NOT EXISTS forum_user
(
    id     bigserial                     NOT NULL PRIMARY KEY,
    "user" bigint REFERENCES "user" (id) NOT NULL,
    forum  bigint REFERENCES forum (id)  NOT NULL
);

CREATE UNLOGGED TABLE IF NOT EXISTS thread
(
    id      bigserial NOT NULL PRIMARY KEY,
    title   text      NOT NULL,
    author  citext    NOT NULL,
    forum   citext,
    message text      NOT NULL,
    votes   int         DEFAULT 0,
    slug    citext,
    created timestamptz DEFAULT now()
);

CREATE UNLOGGED TABLE IF NOT EXISTS post
(
    id        bigserial NOT NULL PRIMARY KEY,
    parent    bigint             DEFAULT 0,
    author    citext    NOT NULL,
    message   text      NOT NULL,
    is_edited bool               DEFAULT false,
    forum     citext,
    thread    int,
    created   timestamptz        DEFAULT now(),
    path      bigint[]  NOT NULL DEFAULT '{0}'
);

CREATE UNLOGGED TABLE IF NOT EXISTS vote
(
    id     bigserial                     NOT NULL PRIMARY KEY,
    "user" bigint REFERENCES "user" (id) NOT NULL,
    thread bigint REFERENCES thread (id) NOT NULL,
    voice  int,
    CONSTRAINT checks UNIQUE ("user", thread)
);

CREATE OR REPLACE FUNCTION thread_vote() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE "thread"
    SET "votes"=(votes + new.voice)
    WHERE "id" = new.thread;
    RETURN new;
end;
$$ language plpgsql;

CREATE TRIGGER "vote_insert"
    AFTER INSERT
    ON "vote"
    FOR EACH ROW
EXECUTE PROCEDURE thread_vote();

CREATE OR REPLACE FUNCTION thread_vote_UPDATE() RETURNS TRIGGER AS
$$
BEGIN
    UPDATE "thread"
    SET "votes"=(votes + 2 * new.voice)
    WHERE "id" = new.thread;
    RETURN new;
END;
$$ language plpgsql;

CREATE TRIGGER "vote_update"
    AFTER UPDATE
    ON "vote"
    FOR EACH ROW
EXECUTE PROCEDURE thread_vote_UPDATE();

CREATE OR REPLACE FUNCTION create_post() RETURNS TRIGGER AS
$$
DECLARE
    _id bigint;

BEGIN
    SELECT u.id, u.nickname, u.fullname, u.about, u.email
    FROM "user" u
    WHERE u.nickname = new.author
    INTO _id;

    UPDATE forum
    SET posts = posts + 1
    WHERE slug = new.forum;
    new.path = (SELECT path FROM post WHERE id = new.parent LIMIT 1) || new.id;
    INSERT INTO forum_user ("user", forum)
    VALUES (_id, (SELECT "id" FROM "forum" WHERE new.forum = slug));
    RETURN new;
END
$$ language plpgsql;

CREATE TRIGGER create_post
    BEFORE INSERT
    ON post
    FOR EACH ROW
EXECUTE PROCEDURE create_post();

CREATE OR REPLACE FUNCTION create_thread() RETURNS TRIGGER AS
$$
DECLARE
    _id bigint;

BEGIN
    SELECT u.id
    FROM "user" u
    WHERE u.nickname = new.author
    INTO _id;

    UPDATE forum
    SET threads = threads + 1
    WHERE slug = new.forum;
    INSERT INTO forum_user ("user", forum)
    VALUES (_id, (SELECT "id" FROM "forum" WHERE new.forum = slug));
    RETURN new;
END
$$ language plpgsql;

CREATE TRIGGER create_thread
    BEFORE INSERT
    ON thread
    FOR EACH ROW
EXECUTE PROCEDURE create_thread();

DROP INDEX IF EXISTS user_nickname_idx;
CREATE INDEX IF NOT EXISTS user_nickname_idx ON "user" (nickname);
DROP INDEX IF EXISTS user_info_idx;
CREATE INDEX IF NOT EXISTS user_info_idx on "user" (nickname, fullname, about, email);

DROP INDEX IF EXISTS forum_slug_idx;
CREATE INDEX IF NOT EXISTS forum_slug_idx ON forum ("slug");
DROP INDEX IF EXISTS forum_user_idx;
CREATE INDEX IF NOT EXISTS forum_user_idx ON forum ("user");

DROP INDEX IF EXISTS forum_user_idx;
CREATE INDEX IF NOT EXISTS forum_user_idx ON forum_user (forum, "user");

DROP INDEX IF EXISTS post_thread_idx;
CREATE INDEX IF NOT EXISTS post_thread_idx ON post (thread);
DROP INDEX IF EXISTS  post_thread_path_idx;
CREATE INDEX IF NOT EXISTS post_thread_path_idx ON post (thread, path);
DROP INDEX IF EXISTS  post_path_parent_idx;
CREATE INDEX IF NOT EXISTS post_path_parent_idx ON post (thread, id, (path[1]), parent);

DROP INDEX IF EXISTS thread_slug_idx;
CREATE INDEX IF NOT EXISTS thread_slug_idx ON thread (slug);
DROP INDEX IF EXISTS thread_author_idx;
CREATE INDEX IF NOT EXISTS thread_author_idx ON thread (author);
DROP INDEX IF EXISTS thread_forum_idx;
CREATE INDEX IF NOT EXISTS thread_forum_idx ON thread (forum);
DROP INDEX IF EXISTS thread_created_idx;
CREATE INDEX IF NOT EXISTS thread_created_idx ON thread (created);

DROP INDEX IF EXISTS vote_user_thread_idx;
CREATE INDEX IF NOT EXISTS vote_user_thread_idx ON vote ("user", thread);

VACUUM ANALYSE;
