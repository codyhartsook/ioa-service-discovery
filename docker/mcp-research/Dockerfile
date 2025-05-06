# Use an official Python base image
FROM python:3.11-slim

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    POETRY_VIRTUALENVS_CREATE=false

# Install uv (if using Astral's uv as package manager)
RUN pip install uv

# Set workdir
WORKDIR /app

# Copy project files
COPY . .

# Expose the port used by your app
EXPOSE 8000

# Set default command
CMD ["uv", "run", "mcp-simple-tool", "--transport", "sse", "--port", "8000"]