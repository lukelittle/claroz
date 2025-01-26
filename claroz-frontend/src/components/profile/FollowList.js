import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  Avatar,
  IconButton,
  Typography,
  Box,
} from '@mui/material';
import { Close as CloseIcon } from '@mui/icons-material';
import { Link as RouterLink } from 'react-router-dom';

function FollowList({ open, onClose, title, users }) {
  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h6">{title}</Typography>
          <IconButton edge="end" onClick={onClose} aria-label="close">
            <CloseIcon />
          </IconButton>
        </Box>
      </DialogTitle>
      <DialogContent dividers>
        <List sx={{ pt: 0 }}>
          {users.map((user) => (
            <ListItem
              key={user.id}
              component={RouterLink}
              to={`/profile/${user.id}`}
              onClick={onClose}
              button
              sx={{
                '&:hover': {
                  backgroundColor: 'action.hover',
                },
              }}
            >
              <ListItemAvatar>
                <Avatar src={user.avatar} alt={user.username}>
                  {user.username[0].toUpperCase()}
                </Avatar>
              </ListItemAvatar>
              <ListItemText
                primary={user.fullName}
                secondary={
                  <>
                    @{user.username}
                    {user.federationType === 'remote' && ` (${user.handle})`}
                  </>
                }
              />
            </ListItem>
          ))}
          {users.length === 0 && (
            <ListItem>
              <ListItemText
                primary={
                  <Typography color="text.secondary" align="center">
                    No users to display
                  </Typography>
                }
              />
            </ListItem>
          )}
        </List>
      </DialogContent>
    </Dialog>
  );
}

export default FollowList;
