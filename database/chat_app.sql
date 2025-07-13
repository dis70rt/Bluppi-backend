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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: conversation_participants; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.conversation_participants (
    conversation_id uuid NOT NULL,
    user_id text NOT NULL,
    joined_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.conversation_participants OWNER TO ethernode;

--
-- Name: conversations; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.conversations (
    conversation_id uuid NOT NULL,
    conversation_name text,
    is_group boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.conversations OWNER TO ethernode;

--
-- Name: message_status; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.message_status (
    message_id uuid NOT NULL,
    user_id text NOT NULL,
    status text,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.message_status OWNER TO ethernode;

--
-- Name: messages; Type: TABLE; Schema: public; Owner: ethernode
--

CREATE TABLE public.messages (
    message_id uuid NOT NULL,
    conversation_id uuid,
    sender_id text,
    message_text text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    type text DEFAULT 'text'::text NOT NULL
);


ALTER TABLE public.messages OWNER TO ethernode;

--
-- Name: conversation_participants conversation_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.conversation_participants
    ADD CONSTRAINT conversation_participants_pkey PRIMARY KEY (conversation_id, user_id);


--
-- Name: conversations conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_pkey PRIMARY KEY (conversation_id);


--
-- Name: message_status message_status_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.message_status
    ADD CONSTRAINT message_status_pkey PRIMARY KEY (message_id, user_id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (message_id);


--
-- Name: idx_conversation_participants_conversation_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_conversation_participants_conversation_id ON public.conversation_participants USING btree (conversation_id);


--
-- Name: idx_conversation_participants_user_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_conversation_participants_user_id ON public.conversation_participants USING btree (user_id);


--
-- Name: idx_conversations_conversation_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_conversations_conversation_id ON public.conversations USING btree (conversation_id);


--
-- Name: idx_messages_conversation_id; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_messages_conversation_id ON public.messages USING btree (conversation_id);


--
-- Name: idx_messages_conversation_id_created_at; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_messages_conversation_id_created_at ON public.messages USING btree (conversation_id, created_at DESC);


--
-- Name: idx_messages_created_at; Type: INDEX; Schema: public; Owner: ethernode
--

CREATE INDEX idx_messages_created_at ON public.messages USING btree (created_at);


--
-- Name: conversation_participants conversation_participants_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.conversation_participants
    ADD CONSTRAINT conversation_participants_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.conversations(conversation_id);


--
-- Name: messages messages_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: ethernode
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.conversations(conversation_id);


--
-- PostgreSQL database dump complete
--

