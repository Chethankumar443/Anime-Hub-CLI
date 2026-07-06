if (typeof global.File === 'undefined') {
  global.File = require('buffer').File;
}

const express = require('express');
const { ANIME, META } = require('@consumet/extensions');

const app = express();
const port = process.env.PORT || 3000;

// Initialize providers
const animeProvider = new ANIME.Hianime();
const anilist = new META.Anilist();

// Root endpoint / health check
app.get('/', (req, res) => {
  res.json({ message: "Welcome to the Consumet API!" });
});

// Gogoanime Wrapper Endpoints (mapped to Hianime for stability)
app.get('/anime/gogoanime/:query', async (req, res) => {
  try {
    const query = req.params.query;
    const page = req.query.page || 1;
    const results = await animeProvider.search(query, page);
    res.json(results);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get('/anime/gogoanime/info/:id', async (req, res) => {
  try {
    const id = req.params.id;
    const info = await animeProvider.fetchAnimeInfo(id);
    res.json(info);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get('/anime/gogoanime/watch/:episodeId', async (req, res) => {
  try {
    const episodeId = req.params.episodeId;
    const sources = await animeProvider.fetchEpisodeSources(episodeId);
    res.json(sources);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// AniList Provider Endpoints
app.get('/meta/anilist/:query', async (req, res) => {
  try {
    const query = req.params.query;
    const page = req.query.page || 1;
    const perPage = req.query.perPage || 10;
    const results = await anilist.search(query, page, perPage);
    res.json(results);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get('/meta/anilist/info/:id', async (req, res) => {
  try {
    const id = req.params.id;
    const info = await anilist.fetchAnimeInfo(id);
    res.json(info);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.get('/meta/anilist/watch/:episodeId', async (req, res) => {
  try {
    const episodeId = req.params.episodeId;
    const sources = await anilist.fetchEpisodeSources(episodeId);
    res.json(sources);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// Bind only to 127.0.0.1 to avoid Windows Firewall network-access prompts
const host = '127.0.0.1';
app.listen(port, host, () => {
  console.log(`Consumet API server listening at http://${host}:${port}`);
});
