# Groupie Trackers

![alt text](image.png)

## Overview

Groupie Trackers is a web application that manipulates and displays data from a given API about bands and artists. The project focuses on creating a user-friendly interface to visualize information about artists, their concert locations, dates, and the relationships between these data points.

## Features

- Displaying artist information including names, images, formation year, first album date, and members
- Showing last and upcoming concert locations
- Present last and upcoming concert dates
- Visualize the relationships between artists, dates, and locations
- Implement client-server interactions for dynamic data retrieval

## API Structure

The API consists of four main parts:

1. **Artists**: Contains basic information about bands and artists
2. **Locations**: Provides data about concert locations
3. **Dates**: Includes information about concert dates
4. **Relation**: Links the artists, dates, and locations data

## Setup and Installation

### File Structure

```bash
groupie-tracker
├─ LICENSE
├─ README.md
├─ cmd
│  └─ server
│     └─ main.go
├─ go.mod
├─ image.png
├─ internal
│  ├─ cache
│  │  └─ cache.go
│  ├─ handlers
│  │  ├─ artists.go
│  │  ├─ fetch.go
│  │  ├─ handlers.go
│  │  ├─ handlers_test.go
│  │  └─ relations.go
│  ├─ models
│  │  └─ models.go
│  └─ templates
│     ├─ artist.html
│     ├─ concerts.html
│     ├─ dates.html
│     ├─ error.html
│     ├─ index.html
│     └─ locations.html
├─ run.sh
└─ static
   └─ css
      └─ style.css

```

### Setup

1. Clone the repository

```go
    git clone https://learn.zone01kisumu.ke/git/skisenge/groupie-tracker.git
```

2. Navigate to root directory and run the script to start the server

```bash
    cd groupie-tracker
    ./run.sh
```

3. Open a web browser and visit `http://localhost:8080`

## Usage

- Browse the main page to see an overview of all artists
- Click on an artist to view detailed information
- Use the search functionality to find specific artists
- Explore the interactive map to see concert locations
- Check the calendar for upcoming concert dates

### Running Tests

To run the unit tests, use the following command:

```bash
    cd internal/handlers
    go test -v
```

## Contributing

1. Fork the repository
2. Create a new branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Create a new Pull Request

## License

[MIT License](LICENSE)

## Acknowledgments

- Thanks to the creators of the Groupie Trackers API for providing the data
- Inspired by the need for an interactive way to explore music artist information

## Authors

[Raymond Ogwel](https://github.com/anxielray)

[Stephen Kisengese](https://github.com/stkisengese)

[Stephen Omotto](https://github.com/somotto)
