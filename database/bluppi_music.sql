--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9 (Ubuntu 16.9-0ubuntu0.24.10.1)
-- Dumped by pg_dump version 17.5 (Ubuntu 17.5-0ubuntu0.25.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
-- SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: event_type; Type: TYPE; Schema: public; Owner: ethernode
--

CREATE TYPE public.event_type AS ENUM (
    'PLAY',
    'PAUSE',
    'SEEK',
    'SKIP'
);


ALTER TYPE public.event_type OWNER TO ethernode;

--
-- Name: playback_status; Type: TYPE; Schema: public; Owner: ethernode
--

CREATE TYPE public.playback_status AS ENUM (
    'PLAYING',
    'PAUSED'
);


ALTER TYPE public.playback_status OWNER TO ethernode;

--
-- Name: room_member_role; Type: TYPE; Schema: public; Owner: ethernode
--

CREATE TYPE public.room_member_role AS ENUM (
    'HOST',
    'MODERATOR',
    'PARTICIPANT'
);


ALTER TYPE public.room_member_role OWNER TO ethernode;

--
-- Name: room_status; Type: TYPE; Schema: public; Owner: ethernode
--

CREATE TYPE public.room_status AS ENUM (
    'ACTIVE',
    'INACTIVE',
    'CLOSED'
);


ALTER TYPE public.room_status OWNER TO ethernode;

--
-- Name: room_visibility; Type: TYPE; Schema: public; Owner: ethernode
--

CREATE TYPE public.room_visibility AS ENUM (
    'PUBLIC',
    'PRIVATE'
);


ALTER TYPE public.room_visibility OWNER TO ethernode;

--
-- Name: track_interaction_type; Type: TYPE; Schema: public; Owner: ethernode
--

CREATE TYPE public.track_interaction_type AS ENUM (
    'liked',
    'last_played',
    'most_played',
    'history'
);


ALTER TYPE public.track_interaction_type OWNER TO ethernode;

--
-- Name: get_active_room_members(text); Type: FUNCTION; Schema: public; Owner: ethernode
--

CREATE FUNCTION public.get_active_room_members(p_room_id text) RETURNS TABLE(user_id text, username text, name text, avatar_url text, role public.room_member_role, joined_at timestamp with time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        u.id,
        u.username,
        u.name,
        u.profile_pic,
        rm.role,
        rm.joined_at
    FROM room_members rm
    JOIN users u ON rm.user_id = u.id
    WHERE rm.room_id = p_room_id 
    AND rm.left_at IS NULL
    ORDER BY rm.joined_at;
END;
$$;


ALTER FUNCTION public.get_active_room_members(p_room_id text) OWNER TO ethernode;

--
-- Name: get_room_queue(text); Type: FUNCTION; Schema: public; Owner: ethernode
--

CREATE FUNCTION public.get_room_queue(p_room_id text) RETURNS TABLE(track_index integer, track_id integer, title character varying, artist character varying, image_url character varying, duration integer, added_by_username text, added_at timestamp with time zone)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        rq.position,
        t.id,
        t.title,
        t.artist,
        t.image_url,
        t.duration,
        u.username,
        rq.added_at
    FROM room_queue rq
    JOIN tracks t ON rq.track_id = t.id
    JOIN users u ON rq.added_by = u.id
    WHERE rq.room_id = p_room_id
    ORDER BY rq.position;
END;
$$;


ALTER FUNCTION public.get_room_queue(p_room_id text) OWNER TO ethernode;

--
-- Name: trg_update_follow_counts(); Type: FUNCTION; Schema: public; Owner: ethernode
--

CREATE FUNCTION public.trg_update_follow_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  IF (TG_OP = 'INSERT') THEN
    UPDATE users SET following_count = following_count + 1 WHERE id = NEW.follower_id;
    UPDATE users SET follower_count = follower_count + 1 WHERE id = NEW.followee_id;
    RETURN NEW;
  ELSIF (TG_OP = 'DELETE') THEN
    UPDATE users SET following_count = following_count - 1 WHERE id = OLD.follower_id;
    UPDATE users SET follower_count = follower_count - 1 WHERE id = OLD.followee_id;
    RETURN OLD;
  END IF;
  RETURN NULL;
END;
$$;


ALTER FUNCTION public.trg_update_follow_counts() OWNER TO ethernode;

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: ethernode
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO ethernode;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: follows; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.follows (
    follower_id text NOT NULL,
    followee_id text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.follows OWNER TO ethernode;

--
-- Name: history_tracks; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.history_tracks (
    id bigint NOT NULL,
    user_id text NOT NULL,
    track_id uuid NOT NULL,
    played_at timestamp with time zone NOT NULL
);


ALTER TABLE public.history_tracks OWNER TO ethernode;

--
-- Name: history_tracks_id_seq; Type: SEQUENCE; Schema: public; Owner: ethernode
--

CREATE SEQUENCE public.history_tracks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.history_tracks_id_seq OWNER TO ethernode;

--
-- Name: history_tracks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: ethernode
--

ALTER SEQUENCE public.history_tracks_id_seq OWNED BY public.history_tracks.id;


--
-- Name: playback_event_log; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.playback_event_log (
    event_id bigint NOT NULL,
    room_id text NOT NULL,
    user_id text NOT NULL,
    event_type public.event_type NOT NULL,
    event_payload jsonb,
    "timestamp" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.playback_event_log OWNER TO ethernode;

--
-- Name: playback_event_log_event_id_seq; Type: SEQUENCE; Schema: public; Owner: ethernode
--

CREATE SEQUENCE public.playback_event_log_event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.playback_event_log_event_id_seq OWNER TO ethernode;

--
-- Name: playback_event_log_event_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: ethernode
--

ALTER SEQUENCE public.playback_event_log_event_id_seq OWNED BY public.playback_event_log.event_id;


--
-- Name: playback_state; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.playback_state (
    room_id text NOT NULL,
    current_track_id uuid,
    position_ms integer DEFAULT 0,
    status public.playback_status DEFAULT 'PAUSED'::public.playback_status,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT playback_state_position_ms_check CHECK ((position_ms >= 0))
);


ALTER TABLE public.playback_state OWNER TO ethernode;

--
-- Name: recent_searches; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.recent_searches (
    id bigint NOT NULL,
    user_id text NOT NULL,
    query text NOT NULL,
    searched_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.recent_searches OWNER TO ethernode;

--
-- Name: recent_searches_id_seq; Type: SEQUENCE; Schema: public; Owner: ethernode
--

CREATE SEQUENCE public.recent_searches_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.recent_searches_id_seq OWNER TO ethernode;

--
-- Name: recent_searches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: ethernode
--

ALTER SEQUENCE public.recent_searches_id_seq OWNED BY public.recent_searches.id;


--
-- Name: room_members; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.room_members (
    id text NOT NULL,
    room_id text NOT NULL,
    user_id text NOT NULL,
    role public.room_member_role DEFAULT 'PARTICIPANT'::public.room_member_role,
    invited_by text,
    joined_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    left_at timestamp with time zone,
    CONSTRAINT check_left_after_joined CHECK (((left_at IS NULL) OR (left_at >= joined_at)))
);


ALTER TABLE public.room_members OWNER TO ethernode;

--
-- Name: room_queue; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.room_queue (
    id text NOT NULL,
    room_id text NOT NULL,
    "position" integer NOT NULL,
    track_id uuid NOT NULL,
    added_by text NOT NULL,
    added_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_positive_position CHECK (("position" > 0))
);


ALTER TABLE public.room_queue OWNER TO ethernode;

--
-- Name: rooms; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.rooms (
    id text NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    host_user_id text NOT NULL,
    room_code character varying(7) NOT NULL,
    status public.room_status DEFAULT 'ACTIVE'::public.room_status NOT NULL,
    visibility public.room_visibility DEFAULT 'PUBLIC'::public.room_visibility NOT NULL,
    max_members integer DEFAULT 50,
    current_track_id uuid,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.rooms OWNER TO ethernode;

--
-- Name: tracks; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.tracks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title character varying(255) NOT NULL,
    artist character varying(255) NOT NULL,
    album character varying(255),
    duration integer,
    genre text[],
    image_url character varying(255),
    preview_url character varying(255),
    video_id character varying(255),
    listeners integer DEFAULT 0,
    play_count integer DEFAULT 0,
    popularity integer DEFAULT 0
);


ALTER TABLE public.tracks OWNER TO ethernode;

--
-- Name: user_track; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.user_track (
    user_id text NOT NULL,
    track_id uuid NOT NULL,
    interaction_type public.track_interaction_type NOT NULL,
    interacted_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.user_track OWNER TO ethernode;

--
-- Name: users; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.users (
    id text NOT NULL,
    email text NOT NULL,
    username text NOT NULL,
    name text NOT NULL,
    bio text,
    country text,
    phone text,
    profile_pic text,
    favorite_genres text[] DEFAULT '{}'::text[] NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    follower_count integer DEFAULT 0 NOT NULL,
    following_count integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.users OWNER TO ethernode;

--
-- Name: history_tracks id; Type: DEFAULT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.history_tracks ALTER COLUMN id SET DEFAULT nextval('public.history_tracks_id_seq'::regclass);


--
-- Name: playback_event_log event_id; Type: DEFAULT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.playback_event_log ALTER COLUMN event_id SET DEFAULT nextval('public.playback_event_log_event_id_seq'::regclass);


--
-- Name: recent_searches id; Type: DEFAULT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.recent_searches ALTER COLUMN id SET DEFAULT nextval('public.recent_searches_id_seq'::regclass);


--
-- Name: follows follows_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_pkey PRIMARY KEY (follower_id, followee_id);


--
-- Name: history_tracks history_tracks_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.history_tracks
    ADD CONSTRAINT history_tracks_pkey PRIMARY KEY (id);


--
-- Name: playback_event_log playback_event_log_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.playback_event_log
    ADD CONSTRAINT playback_event_log_pkey PRIMARY KEY (event_id);


--
-- Name: playback_state playback_state_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.playback_state
    ADD CONSTRAINT playback_state_pkey PRIMARY KEY (room_id);


--
-- Name: recent_searches recent_searches_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.recent_searches
    ADD CONSTRAINT recent_searches_pkey PRIMARY KEY (id);


--
-- Name: room_members room_members_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_members
    ADD CONSTRAINT room_members_pkey PRIMARY KEY (id);


--
-- Name: room_members room_members_room_id_user_id_joined_at_key; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_members
    ADD CONSTRAINT room_members_room_id_user_id_joined_at_key UNIQUE (room_id, user_id, joined_at);


--
-- Name: room_queue room_queue_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_queue
    ADD CONSTRAINT room_queue_pkey PRIMARY KEY (id);


--
-- Name: room_queue room_queue_room_id_position_key; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_queue
    ADD CONSTRAINT room_queue_room_id_position_key UNIQUE (room_id, "position");


--
-- Name: rooms rooms_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_pkey PRIMARY KEY (id);


--
-- Name: rooms rooms_room_code_key; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_room_code_key UNIQUE (room_code);


--
-- Name: tracks tracks_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.tracks
    ADD CONSTRAINT tracks_pkey PRIMARY KEY (id);


--
-- Name: user_track user_track_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.user_track
    ADD CONSTRAINT user_track_pkey PRIMARY KEY (user_id, track_id, interaction_type);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_follows_by_followee; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_follows_by_followee ON public.follows USING btree (followee_id, created_at DESC);


--
-- Name: idx_follows_by_follower; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_follows_by_follower ON public.follows USING btree (follower_id, created_at DESC);


--
-- Name: idx_history_by_user; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_history_by_user ON public.history_tracks USING btree (user_id, played_at DESC);


--
-- Name: idx_playback_event_log_event_type; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_playback_event_log_event_type ON public.playback_event_log USING btree (event_type);


--
-- Name: idx_playback_event_log_room_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_playback_event_log_room_id ON public.playback_event_log USING btree (room_id);


--
-- Name: idx_playback_event_log_room_timestamp; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_playback_event_log_room_timestamp ON public.playback_event_log USING btree (room_id, "timestamp" DESC);


--
-- Name: idx_playback_event_log_timestamp; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_playback_event_log_timestamp ON public.playback_event_log USING btree ("timestamp" DESC);


--
-- Name: idx_playback_event_log_user_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_playback_event_log_user_id ON public.playback_event_log USING btree (user_id);


--
-- Name: idx_room_members_active; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_members_active ON public.room_members USING btree (room_id, user_id) WHERE (left_at IS NULL);


--
-- Name: idx_room_members_joined_at; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_members_joined_at ON public.room_members USING btree (joined_at DESC);


--
-- Name: idx_room_members_role; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_members_role ON public.room_members USING btree (role);


--
-- Name: idx_room_members_room_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_members_room_id ON public.room_members USING btree (room_id);


--
-- Name: idx_room_members_user_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_members_user_id ON public.room_members USING btree (user_id);


--
-- Name: idx_room_queue_added_at; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_queue_added_at ON public.room_queue USING btree (added_at DESC);


--
-- Name: idx_room_queue_added_by; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_queue_added_by ON public.room_queue USING btree (added_by);


--
-- Name: idx_room_queue_position; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_queue_position ON public.room_queue USING btree (room_id, "position");


--
-- Name: idx_room_queue_room_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_room_queue_room_id ON public.room_queue USING btree (room_id);


--
-- Name: idx_rooms_active; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_rooms_active ON public.rooms USING btree (id, name) WHERE (status = 'ACTIVE'::public.room_status);


--
-- Name: idx_rooms_created_at; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_rooms_created_at ON public.rooms USING btree (created_at DESC);


--
-- Name: idx_rooms_host_user_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_rooms_host_user_id ON public.rooms USING btree (host_user_id);


--
-- Name: idx_rooms_public_active; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_rooms_public_active ON public.rooms USING btree (visibility, status, created_at DESC) WHERE ((visibility = 'PUBLIC'::public.room_visibility) AND (status = 'ACTIVE'::public.room_status));


--
-- Name: idx_rooms_status; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_rooms_status ON public.rooms USING btree (status);


--
-- Name: idx_rooms_status_visibility; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_rooms_status_visibility ON public.rooms USING btree (status, visibility);


--
-- Name: idx_rooms_visibility; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_rooms_visibility ON public.rooms USING btree (visibility);


--
-- Name: idx_searches_query_fts; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_searches_query_fts ON public.recent_searches USING gin (to_tsvector('simple'::regconfig, query));


--
-- Name: users_email_ci; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE UNIQUE INDEX users_email_ci ON public.users USING btree (lower(email));


--
-- Name: users_username_ci; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE UNIQUE INDEX users_username_ci ON public.users USING btree (lower(username));


--
-- Name: follows trg_follows_counts; Type: TRIGGER; Schema: public; Owner: ethernode
--

CREATE TRIGGER trg_follows_counts AFTER INSERT OR DELETE ON public.follows FOR EACH ROW EXECUTE FUNCTION public.trg_update_follow_counts();


--
-- Name: playback_state update_playback_state_updated_at; Type: TRIGGER; Schema: public; Owner: ethernode
--

CREATE TRIGGER update_playback_state_updated_at BEFORE UPDATE ON public.playback_state FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: rooms update_rooms_updated_at; Type: TRIGGER; Schema: public; Owner: ethernode
--

CREATE TRIGGER update_rooms_updated_at BEFORE UPDATE ON public.rooms FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: follows follows_followee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_followee_id_fkey FOREIGN KEY (followee_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: follows follows_follower_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.follows
    ADD CONSTRAINT follows_follower_id_fkey FOREIGN KEY (follower_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: history_tracks history_tracks_track_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.history_tracks
    ADD CONSTRAINT history_tracks_track_id_fkey FOREIGN KEY (track_id) REFERENCES public.tracks(id) ON DELETE CASCADE;


--
-- Name: history_tracks history_tracks_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.history_tracks
    ADD CONSTRAINT history_tracks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: playback_event_log playback_event_log_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.playback_event_log
    ADD CONSTRAINT playback_event_log_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(id) ON DELETE CASCADE;


--
-- Name: playback_event_log playback_event_log_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.playback_event_log
    ADD CONSTRAINT playback_event_log_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: playback_state playback_state_current_track_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.playback_state
    ADD CONSTRAINT playback_state_current_track_id_fkey FOREIGN KEY (current_track_id) REFERENCES public.tracks(id) ON DELETE SET NULL;


--
-- Name: playback_state playback_state_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.playback_state
    ADD CONSTRAINT playback_state_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(id) ON DELETE CASCADE;


--
-- Name: recent_searches recent_searches_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.recent_searches
    ADD CONSTRAINT recent_searches_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: room_members room_members_invited_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_members
    ADD CONSTRAINT room_members_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: room_members room_members_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_members
    ADD CONSTRAINT room_members_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(id) ON DELETE CASCADE;


--
-- Name: room_members room_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_members
    ADD CONSTRAINT room_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: room_queue room_queue_added_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_queue
    ADD CONSTRAINT room_queue_added_by_fkey FOREIGN KEY (added_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: room_queue room_queue_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_queue
    ADD CONSTRAINT room_queue_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(id) ON DELETE CASCADE;


--
-- Name: room_queue room_queue_track_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.room_queue
    ADD CONSTRAINT room_queue_track_id_fkey FOREIGN KEY (track_id) REFERENCES public.tracks(id) ON DELETE CASCADE;


--
-- Name: rooms rooms_current_track_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_current_track_id_fkey FOREIGN KEY (current_track_id) REFERENCES public.tracks(id);


--
-- Name: rooms rooms_host_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_host_user_id_fkey FOREIGN KEY (host_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_track user_track_track_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.user_track
    ADD CONSTRAINT user_track_track_id_fkey FOREIGN KEY (track_id) REFERENCES public.tracks(id) ON DELETE CASCADE;


--
-- Name: user_track user_track_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.user_track
    ADD CONSTRAINT user_track_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

