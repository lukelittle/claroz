import React, { useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import {
  Card,
  CardContent,
  CardActions,
  Typography,
  IconButton,
  Button,
  Box,
  TextField,
  Avatar,
  Link,
  Collapse,
  CircularProgress,
} from '@mui/material';
import {
  Favorite as FavoriteIcon,
  FavoriteBorder as FavoriteBorderIcon,
  Comment as CommentIcon,
} from '@mui/icons-material';
import { postAPI } from '../../services/api';

function PostCard({ post, onPostUpdate }) {
  const [isLiked, setIsLiked] = useState(post.isLiked);
  const [likesCount, setLikesCount] = useState(post.likesCount || 0);
  const [showComments, setShowComments] = useState(false);
  const [newComment, setNewComment] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleLikeToggle = async () => {
    try {
      setLoading(true);
      if (isLiked) {
        await postAPI.unlikePost(post.id);
        setLikesCount((prev) => prev - 1);
      } else {
        await postAPI.likePost(post.id);
        setLikesCount((prev) => prev + 1);
      }
      setIsLiked(!isLiked);
    } catch (err) {
      setError('Failed to update like status');
      console.error('Error updating like:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleAddComment = async (e) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    try {
      setLoading(true);
      await postAPI.addComment(post.id, { content: newComment });
      setNewComment('');
      if (onPostUpdate) {
        onPostUpdate(post.id);
      }
    } catch (err) {
      setError('Failed to add comment');
      console.error('Error adding comment:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card sx={{ 
      mb: 2,
      maxWidth: { xs: '100%', sm: '600px' },
      mx: 'auto',
      transition: 'transform 0.2s ease-in-out',
      '&:hover': {
        transform: 'translateY(-2px)',
      }
    }}>
      <CardContent sx={{ p: { xs: 2, sm: 3 } }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Avatar
            src={post.user.avatar}
            alt={post.user.username}
            sx={{ 
              mr: 2,
              width: 48,
              height: 48,
              border: 1,
              borderColor: 'primary.light'
            }}
          />
          <Box>
            <Link
              component={RouterLink}
              to={`/profile/${post.user.id}`}
              color="inherit"
              underline="hover"
            >
              <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                {post.user.username}
                {post.user.federationType === 'remote' &&
                  ` (${post.user.handle})`}
              </Typography>
            </Link>
            <Typography variant="caption" color="text.secondary">
              {new Date(post.createdAt).toLocaleString()}
            </Typography>
          </Box>
        </Box>

        <Typography variant="body1" sx={{ 
          mb: 2,
          lineHeight: 1.6,
          whiteSpace: 'pre-wrap'
        }}>
          {post.content}
        </Typography>

        {error && (
          <Typography color="error" variant="body2" sx={{ mb: 1 }}>
            {error}
          </Typography>
        )}
      </CardContent>

      <CardActions disableSpacing sx={{ px: { xs: 2, sm: 3 }, py: 1 }}>
          <IconButton
            onClick={handleLikeToggle}
            disabled={loading}
            color={isLiked ? 'primary' : 'default'}
            sx={{
              '&:hover': {
                backgroundColor: 'rgba(25, 118, 210, 0.04)'
              }
            }}
          >
          {loading ? (
            <CircularProgress size={24} />
          ) : isLiked ? (
            <FavoriteIcon />
          ) : (
            <FavoriteBorderIcon />
          )}
        </IconButton>
        <Typography variant="body2" color="text.secondary" sx={{ mr: 2 }}>
          {likesCount}
        </Typography>

        <IconButton onClick={() => setShowComments(!showComments)}>
          <CommentIcon />
        </IconButton>
        <Typography variant="body2" color="text.secondary">
          {post.comments?.length || 0}
        </Typography>
      </CardActions>

      <Collapse in={showComments} timeout="auto" unmountOnExit>
        <CardContent>
          <form onSubmit={handleAddComment}>
            <TextField
              fullWidth
              size="small"
              placeholder="Add a comment..."
              value={newComment}
              onChange={(e) => setNewComment(e.target.value)}
              disabled={loading}
              sx={{ mb: 2 }}
            />
            <Button
              type="submit"
              variant="contained"
              size="small"
              disabled={!newComment.trim() || loading}
            >
              {loading ? <CircularProgress size={24} /> : 'Comment'}
            </Button>
          </form>

          <Box sx={{ mt: 3, pl: 1 }}>
            {post.comments?.map((comment) => (
              <Box 
                key={comment.id} 
                sx={{ 
                  mb: 2,
                  p: 1.5,
                  borderRadius: 1,
                  bgcolor: 'background.default'
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 0.5 }}>
                  <Avatar
                    src={comment.user.avatar}
                    alt={comment.user.username}
                    sx={{ 
                      width: 32,
                      height: 32,
                      mr: 1,
                      border: 1,
                      borderColor: 'primary.light'
                    }}
                  />
                  <Link
                    component={RouterLink}
                    to={`/profile/${comment.user.id}`}
                    color="inherit"
                    underline="hover"
                  >
                    <Typography variant="subtitle2">
                      {comment.user.username}
                    </Typography>
                  </Link>
                </Box>
                <Typography variant="body2">{comment.content}</Typography>
                <Typography variant="caption" color="text.secondary">
                  {new Date(comment.createdAt).toLocaleString()}
                </Typography>
              </Box>
            ))}
          </Box>
        </CardContent>
      </Collapse>
    </Card>
  );
}

export default PostCard;
