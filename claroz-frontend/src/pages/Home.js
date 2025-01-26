import React, { useState, useEffect } from 'react';
import {
  Container,
  Grid,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Box,
  CircularProgress,
  Alert,
} from '@mui/material';
import Layout from '../components/shared/Layout';
import { postAPI, federationAPI } from '../services/api';
import { useAuth } from '../context/AuthContext';

function Home() {
  useAuth(); // Keep auth context for protected route
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [newPost, setNewPost] = useState('');
  const [federatedHandle, setFederatedHandle] = useState('');
  const [federationLoading, setFederationLoading] = useState(false);

  useEffect(() => {
    fetchPosts();
  }, []);

  const fetchPosts = async () => {
    try {
      setLoading(true);
      const response = await postAPI.getPosts();
      setPosts(response.data);
    } catch (err) {
      setError('Failed to fetch posts');
      console.error('Error fetching posts:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreatePost = async (e) => {
    e.preventDefault();
    if (!newPost.trim()) return;

    try {
      const response = await postAPI.createPost({ content: newPost });
      setPosts([response.data, ...posts]);
      setNewPost('');
    } catch (err) {
      setError('Failed to create post');
      console.error('Error creating post:', err);
    }
  };

  const handleResolveFederatedProfile = async (e) => {
    e.preventDefault();
    if (!federatedHandle.trim()) return;

    try {
      setFederationLoading(true);
      await federationAPI.resolveProfile(federatedHandle);
      setFederatedHandle('');
      // Refresh posts to include federated content
      fetchPosts();
    } catch (err) {
      setError('Failed to resolve federated profile');
      console.error('Error resolving federated profile:', err);
    } finally {
      setFederationLoading(false);
    }
  };

  if (loading) {
    return (
      <Layout>
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '50vh',
          }}
        >
          <CircularProgress />
        </Box>
      </Layout>
    );
  }

  return (
    <Layout>
      <Container maxWidth="md">
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Create Post
                </Typography>
                <form onSubmit={handleCreatePost}>
                  <TextField
                    fullWidth
                    multiline
                    rows={3}
                    variant="outlined"
                    placeholder="What's on your mind?"
                    value={newPost}
                    onChange={(e) => setNewPost(e.target.value)}
                    sx={{ mb: 2 }}
                  />
                  <Button
                    type="submit"
                    variant="contained"
                    disabled={!newPost.trim()}
                  >
                    Post
                  </Button>
                </form>
              </CardContent>
            </Card>
          </Grid>

          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Follow Federated Profile
                </Typography>
                <form onSubmit={handleResolveFederatedProfile}>
                  <TextField
                    fullWidth
                    variant="outlined"
                    placeholder="Enter handle (e.g. user.bsky.social)"
                    value={federatedHandle}
                    onChange={(e) => setFederatedHandle(e.target.value)}
                    sx={{ mb: 2 }}
                  />
                  <Button
                    type="submit"
                    variant="contained"
                    disabled={!federatedHandle.trim() || federationLoading}
                  >
                    {federationLoading ? <CircularProgress size={24} /> : 'Resolve'}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </Grid>

          {posts.map((post) => (
            <Grid item xs={12} key={post.id}>
              <Card>
                <CardContent>
                  <Typography variant="subtitle2" color="text.secondary">
                    {post.user.username}
                    {post.user.federationType === 'remote' &&
                      ` (${post.user.handle})`}
                  </Typography>
                  <Typography variant="body1" sx={{ mt: 1 }}>
                    {post.content}
                  </Typography>
                  <Typography
                    variant="caption"
                    color="text.secondary"
                    sx={{ mt: 1, display: 'block' }}
                  >
                    {new Date(post.createdAt).toLocaleString()}
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Container>
    </Layout>
  );
}

export default Home;
