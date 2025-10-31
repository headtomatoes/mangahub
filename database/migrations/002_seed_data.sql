-- ========================================
-- GENRES
-- ========================================
INSERT INTO genres (name) VALUES
('Action'),
('Adventure'),
('Comedy'),
('Drama'),
('Fantasy'),
('Horror'),
('Mystery'),
('Psychological'),
('Romance'),
('Sci-Fi'),
('Slice of Life'),
('Sports'),
('Supernatural'),
('Thriller'),
('Tragedy'),
('Seinen'),
('Shounen'),
('Shoujo'),
('Josei'),
('Isekai'),
('Mecha'),
('Historical'),
('School'),
('Martial Arts'),
('Demons');

-- ========================================
-- MANGA (20 popular titles)
-- ========================================
INSERT INTO manga (slug, title, author, status, total_chapters, description, cover_url) VALUES
('one-piece', 'One Piece', 'Oda Eiichiro', 'ongoing', 1095, 'The story follows Monkey D. Luffy, a young man whose body gained the properties of rubber after unintentionally eating a Devil Fruit.', 'https://example.com/one-piece.jpg'),
('naruto', 'Naruto', 'Kishimoto Masashi', 'completed', 700, 'The story of Naruto Uzumaki, a young ninja who seeks recognition from his peers and dreams of becoming the Hokage.', 'https://example.com/naruto.jpg'),
('attack-on-titan', 'Attack on Titan', 'Isayama Hajime', 'completed', 139, 'Humanity lives within cities surrounded by enormous walls that protect them from gigantic man-eating humanoids.', 'https://example.com/aot.jpg'),
('demon-slayer', 'Demon Slayer: Kimetsu no Yaiba', 'Gotouge Koyoharu', 'completed', 205, 'A family is attacked by demons and only two members survive - Tanjiro and his sister Nezuko, who is slowly turning into a demon.', 'https://example.com/demon-slayer.jpg'),
('my-hero-academia', 'My Hero Academia', 'Horikoshi Kouhei', 'ongoing', 405, 'In a world where nearly everyone has superpowers, Izuku Midoriya dreams of becoming a hero despite being born without a Quirk.', 'https://example.com/mha.jpg'),
('tokyo-ghoul', 'Tokyo Ghoul', 'Ishida Sui', 'completed', 143, 'Ken Kaneki is a college student who barely survives a deadly encounter with Rize Kamishiro, his date who reveals herself as a ghoul.', 'https://example.com/tokyo-ghoul.jpg'),
('death-note', 'Death Note', 'Ohba Tsugumi', 'completed', 108, 'A high school student discovers a supernatural notebook that allows him to kill anyone by writing their name.', 'https://example.com/death-note.jpg'),
('fullmetal-alchemist', 'Fullmetal Alchemist', 'Arakawa Hiromu', 'completed', 116, 'Two brothers search for a Philosopher''s Stone after an attempt to revive their deceased mother goes wrong.', 'https://example.com/fma.jpg'),
('hunter-x-hunter', 'Hunter x Hunter', 'Togashi Yoshihiro', 'hiatus', 390, 'A young boy named Gon Freecss aspires to become a Hunter in order to find his father.', 'https://example.com/hxh.jpg'),
('bleach', 'Bleach', 'Kubo Tite', 'completed', 686, 'Ichigo Kurosaki obtains the powers of a Soul Reaper and must defend humans from evil spirits.', 'https://example.com/bleach.jpg'),
('chainsaw-man', 'Chainsaw Man', 'Fujimoto Tatsuki', 'ongoing', 180, 'Denji has a simple dream - to live a happy and peaceful life, but his reality is harsh.', 'https://example.com/chainsaw-man.jpg'),
('jujutsu-kaisen', 'Jujutsu Kaisen', 'Akutami Gege', 'ongoing', 245, 'A boy swallows a cursed talisman and becomes host to a powerful Curse, gaining supernatural abilities.', 'https://example.com/jjk.jpg'),
('solo-leveling', 'Solo Leveling', 'Chugong', 'completed', 179, 'In a world where hunters fight monsters, the weakest hunter becomes the strongest through a mysterious system.', 'https://example.com/solo-leveling.jpg'),
('vinland-saga', 'Vinland Saga', 'Yukimura Makoto', 'ongoing', 200, 'A young Viking warrior seeks revenge for his father''s murder in medieval Europe.', 'https://example.com/vinland-saga.jpg'),
('berserk', 'Berserk', 'Miura Kentaro', 'hiatus', 364, 'Guts, a former mercenary now branded for death, seeks sanctuary from demons and revenge.', 'https://example.com/berserk.jpg'),
('one-punch-man', 'One Punch Man', 'ONE', 'ongoing', 195, 'The story of Saitama, a hero who can defeat any opponent with a single punch.', 'https://example.com/opm.jpg'),
('tokyo-revengers', 'Tokyo Revengers', 'Wakui Ken', 'completed', 278, 'A young man discovers he can travel back in time and tries to prevent his ex-girlfriend''s death.', 'https://example.com/tokyo-revengers.jpg'),
('vagabond', 'Vagabond', 'Inoue Takehiko', 'hiatus', 327, 'A fictional account of the life of Miyamoto Musashi, one of Japan''s most renowned swordsmen.', 'https://example.com/vagabond.jpg'),
('mob-psycho-100', 'Mob Psycho 100', 'ONE', 'completed', 101, 'A boy with psychic powers tries to live a normal life while facing supernatural threats.', 'https://example.com/mob-psycho.jpg'),
('the-promised-neverland', 'The Promised Neverland', 'Shirai Kaiu', 'completed', 181, 'Children at an orphanage discover the dark truth about their existence and plan their escape.', 'https://example.com/tpn.jpg');

-- ========================================
-- MANGA_GENRES (Associations)
-- ========================================
-- One Piece: Action, Adventure, Comedy, Fantasy, Shounen
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 5), (1, 17);

-- Naruto: Action, Adventure, Martial Arts, Shounen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(2, 1), (2, 2), (2, 24), (2, 17), (2, 13);

-- Attack on Titan: Action, Drama, Fantasy, Horror, Mystery, Shounen
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(3, 1), (3, 4), (3, 5), (3, 6), (3, 7), (3, 17);

-- Demon Slayer: Action, Fantasy, Historical, Shounen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(4, 1), (4, 5), (4, 22), (4, 17), (4, 13);

-- My Hero Academia: Action, Comedy, School, Shounen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(5, 1), (5, 3), (5, 23), (5, 17), (5, 13);

-- Tokyo Ghoul: Action, Drama, Horror, Mystery, Psychological, Seinen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(6, 1), (6, 4), (6, 6), (6, 7), (6, 8), (6, 16), (6, 13);

-- Death Note: Mystery, Psychological, Shounen, Supernatural, Thriller
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(7, 7), (7, 8), (7, 17), (7, 13), (7, 14);

-- Fullmetal Alchemist: Action, Adventure, Drama, Fantasy, Shounen
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(8, 1), (8, 2), (8, 4), (8, 5), (8, 17);

-- Hunter x Hunter: Action, Adventure, Fantasy, Shounen
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(9, 1), (9, 2), (9, 5), (9, 17);

-- Bleach: Action, Adventure, Shounen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(10, 1), (10, 2), (10, 17), (10, 13);

-- Chainsaw Man: Action, Comedy, Horror, Shounen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(11, 1), (11, 3), (11, 6), (11, 17), (11, 13);

-- Jujutsu Kaisen: Action, Fantasy, School, Shounen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(12, 1), (12, 5), (12, 23), (12, 17), (12, 13);

-- Solo Leveling: Action, Adventure, Fantasy
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(13, 1), (13, 2), (13, 5);

-- Vinland Saga: Action, Adventure, Drama, Historical, Seinen
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(14, 1), (14, 2), (14, 4), (14, 22), (14, 16);

-- Berserk: Action, Adventure, Drama, Fantasy, Horror, Seinen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(15, 1), (15, 2), (15, 4), (15, 5), (15, 6), (15, 16), (15, 13);

-- One Punch Man: Action, Comedy, Seinen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(16, 1), (16, 3), (16, 16), (16, 13);

-- Tokyo Revengers: Action, Drama, Shounen, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(17, 1), (17, 4), (17, 17), (17, 13);

-- Vagabond: Action, Adventure, Drama, Historical, Seinen
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(18, 1), (18, 2), (18, 4), (18, 22), (18, 16);

-- Mob Psycho 100: Action, Comedy, Slice of Life, Supernatural
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(19, 1), (19, 3), (19, 11), (19, 13);

-- The Promised Neverland: Mystery, Psychological, Shounen, Thriller
INSERT INTO manga_genres (manga_id, genre_id) VALUES
(20, 7), (20, 8), (20, 17), (20, 14);