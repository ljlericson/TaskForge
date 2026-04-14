<p align="center">
  <img src="web/assets/logo.png" width="400" />
</p>

<p align="center">
  <b>Distributed Worker System for Scalable Job Execution</b>
</p>

<p align="center">
  <img src="https://github.com/ljlericson/TaskForge/actions/workflows/go-test.yml/badge.svg" />
</p>

---

## Overview

**TaskForge** is a distributed system in developement designed to schedule, distribute, and execute jobs across multiple worker nodes efficiently.

It leverages multiple technologies:

- **Golang** for backend scheduling and routing  
- **C++** for high-performance job execution  
- **JavaScript** for frontend job submission  

---

## Workflow

1. Frontend submits a job  
2. Server schedules and queues the job  
3. Idle worker node receives the job  
4. Worker:
   - Downloads required resources  
   - Executes `.jar` file  
   - Uploads results to output endpoint  

---

## Example Use Case

### Large-Scale Data Processing

```java
import fileImporter;
import fileWriter;

class main {
    public static void main(int argc, string argv[]) {
        float values[] = fileImporter.getValuesFromFile("input.txt");
        int size = fileImporter.getNumberOfDataPoints("input.txt");
        
        float sum = 0;
        for(int i = 0; i < size; i++) {
            sum += values[i];
        }

        fileWriter.writeValue("output.txt", sum / size);
    }
}
```

---

## Features

### Current

- Node registry (alive / dead tracking)  
- Event-driven scheduler  
- RSA-based authentication system  
- Binary heap priority queue for jobs  

### Planned

- Sandboxed `.jar` execution  
- Distributed input/output via URLs  
- Full frontend dashboard:
  - Active jobs  
  - Queue state  
  - Job cancellation  

---

## Getting Started

### Clone Repository

```bash
git clone https://github.com/ljlericson/TaskForge/
cd TaskForge
```

---

## Build Instructions

### Server

```bash
sh build-server.sh
./bin/TaskForge-Server
```

---

### Worker (C++)

```bash
sudo apt install openssl cmake python3

cd worker/vcpkg
sh bootstrap_vcpkg.sh
./vcpkg install

cd ../../
sh build-worker.sh
./bin/TaskForge-Client
```
Then for subsequent builds
```bash
sh build-worker.sh
./bin/TaskForge-Client
```

---

### Client

```bash
cd web
python3 -m http.server 8000
```

Open in your browser:  
http://localhost:8000/

---

## Running Tests

```bash
sh test-server.sh
```

---

## Roadmap

- [ ] Secure sandbox execution  
- [ ] Distributed file handling  
- [ ] Horizontal scaling improvements  
- [ ] UI/UX enhancements  
- [ ] Job retry and fault tolerance  

---

## Contributing

Contributions are welcome. Feel free to open issues or submit pull requests.

---

## License

MIT License