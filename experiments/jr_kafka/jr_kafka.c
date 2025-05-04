#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>
#include <librdkafka/rdkafka.h>
#include <errno.h>
#include <signal.h>
#include <ctype.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdbool.h>

#define SOCKET_PATH_FORMAT "/tmp/jr_kafka_%s_socket"
#define FIFO_PATH_FORMAT "/tmp/jr_kafka_%s_fifo"
#define BUFFER_SIZE 1024
#define CONFIG_PATH "./kafka/config.properties"
#define MAX_CONFIG_LINE 1024
#define MAX_CONFIG_PAIRS 50
#define DEFAULT_PORT 8888
#define MAX_KEY_SIZE 256
#define MAX_HEADER_SIZE 1024
#define DELIMITER "|"  // Use Record Separator (␞) character as delimiter
#define DEFAULT_DIR "/tmp"
#define MIN_PORT 1
#define MAX_PORT 65535
#define LOCALHOST "127.0.0.1"

// Add these global variables
static volatile sig_atomic_t running = 1;
static rd_kafka_t *g_producer = NULL;
static int g_server_fd = -1;
static char g_cleanup_path[256] = {0};  // Store path for cleanup

// Structure to hold config key-value pairs
typedef struct {
    char key[256];
    char value[1024];
} ConfigPair;

// Add this structure to store the parsed message parts
typedef struct {
    char key[MAX_KEY_SIZE];
    char header[MAX_HEADER_SIZE];
    char *message;  // Will point to the message part in the buffer
} ParsedMessage;

// Message callback for librdkafka
static void dr_msg_cb(rd_kafka_t*,
                     const rd_kafka_message_t *rkmessage,
                     void*) {
    if (rkmessage->err) {
        fprintf(stderr, "Message delivery failed: %s\n",
                rd_kafka_err2str(rkmessage->err));
    } else {
        fprintf(stderr, "Message delivered to topic %s [%d] at offset %" PRId64 "\n",
                rd_kafka_topic_name(rkmessage->rkt),
                rkmessage->partition,
                rkmessage->offset);
    }
}

// Function to read and parse config file
rd_kafka_conf_t* read_config_file(const char* config_path, char* errstr, size_t errstr_size) {
    FILE* file = fopen(config_path, "r");
    if (!file) {
        snprintf(errstr, errstr_size, "Failed to open config file: %s", config_path);
        return NULL;
    }

    rd_kafka_conf_t* conf = rd_kafka_conf_new();
    char line[MAX_CONFIG_LINE];
    char key[256], value[1024];

    while (fgets(line, sizeof(line), file)) {
        // Skip comments and empty lines
        if (line[0] == '#' || line[0] == '\n') {
            continue;
        }

        // Remove newline
        line[strcspn(line, "\n")] = 0;

        // Parse key=value
        char* separator = strchr(line, '=');
        if (separator) {
            *separator = '\0';
            strncpy(key, line, sizeof(key) - 1);
            strncpy(value, separator + 1, sizeof(value) - 1);

            // Trim whitespace
            char* p = key + strlen(key) - 1;
            while (p >= key && isspace(*p)) *p-- = '\0';
            p = value;
            while (*p && isspace(*p)) p++;
            memmove(value, p, strlen(p) + 1);

            // Set configuration
            rd_kafka_conf_res_t result = rd_kafka_conf_set(conf, key, value,
                                                         errstr, errstr_size);
            if (result != RD_KAFKA_CONF_OK) {
                fprintf(stderr, "Configuration failed for %s=%s: %s\n",
                        key, value, errstr);
                rd_kafka_conf_destroy(conf);
                fclose(file);
                return NULL;
            }
            printf("Configured %s=%s\n", key, value);
        }
    }

    fclose(file);
    return conf;
}

// Add this function before main()
ParsedMessage parse_buffer(char *buffer) {
    ParsedMessage result = {0};  // Initialize all fields to zero/NULL
    char *saveptr;  // For strtok_r thread safety

    // Get key (first part)
    char *token = strtok_r(buffer, DELIMITER, &saveptr);
    if (token) {
        strncpy(result.key, token, MAX_KEY_SIZE - 1);
        result.key[MAX_KEY_SIZE - 1] = '\0';

        // Get header (second part)
        token = strtok_r(NULL, DELIMITER, &saveptr);
        if (token) {
            strncpy(result.header, token, MAX_HEADER_SIZE - 1);
            result.header[MAX_HEADER_SIZE - 1] = '\0';

            // Get message (rest of the buffer)
            token = strtok_r(NULL, "", &saveptr);
            if (token) {
                result.message = token;
            }
        }
    }

    return result;
}

// Add a signal handler function
static void handle_signal(int sig) {
    running = 0;
}

// Add a cleanup function
static void cleanup(void) {
    if (g_producer) {
        rd_kafka_flush(g_producer, 10 * 1000);
        rd_kafka_destroy(g_producer);
    }
    if (g_server_fd != -1) {
        close(g_server_fd);
    }
    if (g_cleanup_path[0] != '\0') {
        unlink(g_cleanup_path);
    }
}

// In main(), modify the initialization code:
int main(const int argc, char **argv) {
    bool use_tcp = false;
    bool use_fifo = false;
    int port = DEFAULT_PORT;
    char fifo_path[256];
    char socket_path[256];
    const char *base_dir = DEFAULT_DIR;
    const char *topic;

    int opt;
    while ((opt = getopt(argc, argv, "tp:fd:")) != -1) {  // Removed 'h' from options
        switch (opt) {
            case 't':
                use_tcp = true;
                break;
            case 'p':
                char *endptr;
                long parsed_port = strtol(optarg, &endptr, 10);
                if (*endptr != '\0' || parsed_port < MIN_PORT || parsed_port > MAX_PORT) {
                    fprintf(stderr, "Invalid port number. Must be between %d and %d\n", MIN_PORT, MAX_PORT);
                    return EXIT_FAILURE;
                }
                port = (int)parsed_port;
                break;
            case 'f':
                use_fifo = true;
                break;
            case 'd':
                base_dir = optarg;
                break;
            default:
                fprintf(stderr, "Usage: %s [-t] [-p port] [-f] [-d directory] <topic>\n"
                        "  -t          Use TCP sockets instead of Unix domain sockets\n"
                        "  -p port     Port number for TCP (default: %d)\n"
                        "  -f          Use named pipe instead of Unix domain socket\n"
                        "  -d dir      Directory for FIFO/socket (default: %s)\n",
                        argv[0], DEFAULT_PORT, DEFAULT_DIR);
                return EXIT_FAILURE;
        }
    }

    if (optind >= argc) {
        fprintf(stderr, "Topic name is required\n");
        return EXIT_FAILURE;
    }
    topic = argv[optind];

    // ... rest of the code, using base_dir for path construction:
    if (use_fifo) {
        snprintf(fifo_path, sizeof(fifo_path), "%s/jr_kafka_%s_fifo", base_dir, topic);
    }
    if (!use_tcp && !use_fifo) {
        snprintf(socket_path, sizeof(socket_path), "%s/jr_kafka_%s_socket", base_dir, topic);
    }

    char errstr[512];
    int server_fd, client_fd;
    char buffer[BUFFER_SIZE];

    // Read configuration and create producer (existing code)...
    rd_kafka_conf_t* conf = read_config_file(CONFIG_PATH, errstr, sizeof(errstr));
    if (!conf) {
        fprintf(stderr, "Failed to read config: %s\n", errstr);
        return EXIT_FAILURE;
    }

    rd_kafka_conf_set_dr_msg_cb(conf, dr_msg_cb);
    rd_kafka_t* producer = rd_kafka_new(RD_KAFKA_PRODUCER, conf, errstr, sizeof(errstr));
    if (!producer) {
        fprintf(stderr, "Failed to create producer: %s\n", errstr);
        return EXIT_FAILURE;
    }

    // Store producer in global variable for cleanup
    g_producer = producer;
    g_server_fd = server_fd;

    if (use_fifo) {
        // Create the FIFO with the topic-specific name
        if (mkfifo(fifo_path, 0666) == -1 && errno != EEXIST) {
            perror("Failed to create FIFO");
            rd_kafka_destroy(producer);
            return EXIT_FAILURE;
        }

        server_fd = open(fifo_path, O_RDONLY);
        if (server_fd == -1) {
            perror("Failed to open FIFO");
            unlink(fifo_path);
            rd_kafka_destroy(producer);
            return EXIT_FAILURE;
        }

        printf("Listening on FIFO: %s\n", fifo_path);
    } else if (use_tcp) {
        // TCP socket setup
        struct sockaddr_in server_addr_in;

        server_fd = socket(AF_INET, SOCK_STREAM, 0);
        if (server_fd == -1) {
            perror("Failed to create socket");
            rd_kafka_destroy(producer);
            return EXIT_FAILURE;
        }

        memset(&server_addr_in, 0, sizeof(server_addr_in));
        server_addr_in.sin_family = AF_INET;
        server_addr_in.sin_addr.s_addr = inet_addr(LOCALHOST);
        server_addr_in.sin_port = htons(port);

        if (bind(server_fd, (struct sockaddr*)&server_addr_in, sizeof(server_addr_in)) == -1) {
            perror("Bind failed");
            close(server_fd);
            rd_kafka_destroy(producer);
            return EXIT_FAILURE;
        }

        printf("Server listening on %s:%d\n", LOCALHOST, port);
    } else {
        // Unix domain socket setup with topic-specific path
        struct sockaddr_un server_addr;

        server_fd = socket(AF_UNIX, SOCK_STREAM, 0);
        if (server_fd == -1) {
            perror("Failed to create socket");
            rd_kafka_destroy(producer);
            return EXIT_FAILURE;
        }

        memset(&server_addr, 0, sizeof(server_addr));
        server_addr.sun_family = AF_UNIX;
        strncpy(server_addr.sun_path, socket_path, sizeof(server_addr.sun_path) - 1);

        unlink(socket_path);  // Remove any existing socket file

        if (bind(server_fd, (struct sockaddr*)&server_addr, sizeof(server_addr)) == -1) {
            perror("Bind failed");
            close(server_fd);
            rd_kafka_destroy(producer);
            return EXIT_FAILURE;
        }

        printf("Server listening on %s\n", socket_path);
    }

    // Store the cleanup path
    if (!use_tcp) {
        if (use_fifo) {
            strncpy(g_cleanup_path, fifo_path, sizeof(g_cleanup_path) - 1);
        } else {
            strncpy(g_cleanup_path, socket_path, sizeof(g_cleanup_path) - 1);
        }
    }

    if (!use_fifo) {
        if (listen(server_fd, 5) == -1) {
            perror("Listen failed");
            close(server_fd);
            if (!use_tcp) {
                unlink(socket_path);
            }
            rd_kafka_destroy(producer);
            return EXIT_FAILURE;
        }
    }

    printf("Kafka producer created. Waiting for messages...\n");

    while (running) {
        if (use_fifo) {
            // Create the FIFO with the topic-specific name
            ssize_t bytes_read = read(server_fd, buffer, BUFFER_SIZE - 1);
            if (bytes_read > 0) {
                buffer[bytes_read] = '\0';
                printf("Received raw data: %s\n", buffer);

                ParsedMessage parsed = parse_buffer(buffer);

                // Validate that we have all parts
                if (!parsed.key[0] || !parsed.header[0] || !parsed.message) {
                    fprintf(stderr, "Invalid message format. Expected: key␞header␞message\n");
                    if (!use_fifo && client_fd != -1) {
                        close(client_fd);
                    }
                    continue;
                }

                printf("Parsed message:\n");
                printf("  Key: %s\n", parsed.key);
                printf("  Header: %s\n", parsed.header);
                printf("  Message: %s\n", parsed.message);

                // Create headers
                rd_kafka_headers_t *headers = rd_kafka_headers_new(1);
                if (headers) {
                    rd_kafka_header_add(headers,
                                      "custom_header", -1,  // -1 means string length will be calculated
                                      parsed.header, -1);
                }

                // Produce message to Kafka with all parts
                rd_kafka_resp_err_t err = rd_kafka_producev(producer,
                                        RD_KAFKA_V_TOPIC(topic),
                                        RD_KAFKA_V_KEY(parsed.key, strlen(parsed.key)),
                                        //RD_KAFKA_V_HEADERS(headers),
                                        RD_KAFKA_V_VALUE(parsed.message, strlen(parsed.message)),
                                        RD_KAFKA_V_MSGFLAGS(RD_KAFKA_MSG_F_COPY),
                                        RD_KAFKA_V_END);

                if (headers) {
                    rd_kafka_headers_destroy(headers);
                }

                if (err) {
                    fprintf(stderr, "Failed to produce message: %s\n",
                            rd_kafka_err2str(err));
                } else {
                    printf("Message queued for delivery\n");
                }

                rd_kafka_flush(producer, 10 * 1000);
            } else if (bytes_read == 0 && running) {
                // Only reopen if we're still supposed to be running
                close(server_fd);
                server_fd = open(fifo_path, O_RDONLY);
                if (server_fd == -1) {
                    perror("Failed to reopen FIFO");
                    break;
                }
                g_server_fd = server_fd;
            } else if (running) {
                perror("Error reading from FIFO");
                break;
            }
        } else {
            // For socket connections, add timeout to accept
            struct timeval tv;
            tv.tv_sec = 1;  // 1 second timeout
            tv.tv_usec = 0;

            fd_set readfds;
            FD_ZERO(&readfds);
            FD_SET(server_fd, &readfds);

            int ret = select(server_fd + 1, &readfds, NULL, NULL, &tv);
            if (ret < 0 && errno != EINTR) {
                perror("Select failed");
                break;
            }
            if (ret > 0) {
                socklen_t client_len;
                struct sockaddr_storage client_addr;
                client_len = sizeof(client_addr);

                client_fd = accept(server_fd, (struct sockaddr*)&client_addr, &client_len);
                if (client_fd == -1) {
                    perror("Accept failed");
                    continue;
                }

                printf("Client connected\n");

                // Rest of the code remains the same...
                ssize_t bytes_read = read(client_fd, buffer, BUFFER_SIZE - 1);
                if (bytes_read > 0) {
                    buffer[bytes_read] = '\0';
                    printf("Received raw data: %s\n", buffer);

                    ParsedMessage parsed = parse_buffer(buffer);

                    // Validate that we have all parts
                    if (!parsed.key[0] || !parsed.header[0] || !parsed.message) {
                        fprintf(stderr, "Invalid message format. Expected: key␞header␞message\n");
                        if (!use_fifo && client_fd != -1) {
                            close(client_fd);
                        }
                        continue;
                    }

                    printf("Parsed message:\n");
                    printf("  Key: %s\n", parsed.key);
                    printf("  Header: %s\n", parsed.header);
                    printf("  Message: %s\n", parsed.message);

                    // Create headers
                    rd_kafka_headers_t *headers = rd_kafka_headers_new(1);
                    if (headers) {
                        rd_kafka_header_add(headers,
                                          "custom_header", -1,  // -1 means string length will be calculated
                                          parsed.header, -1);
                    }

                    // Produce message to Kafka with all parts
                    rd_kafka_resp_err_t err = rd_kafka_producev(producer,
                                            RD_KAFKA_V_TOPIC(topic),
                                            RD_KAFKA_V_KEY(parsed.key, strlen(parsed.key)),
                                            //RD_KAFKA_V_HEADERS(headers),
                                            RD_KAFKA_V_VALUE(parsed.message, strlen(parsed.message)),
                                            RD_KAFKA_V_MSGFLAGS(RD_KAFKA_MSG_F_COPY),
                                            RD_KAFKA_V_END);

                    if (headers) {
                        rd_kafka_headers_destroy(headers);
                    }

                    if (err) {
                        fprintf(stderr, "Failed to produce message: %s\n",
                                rd_kafka_err2str(err));
                    } else {
                        printf("Message queued for delivery\n");
                    }

                    rd_kafka_flush(producer, 10 * 1000);
                }

                close(client_fd);
            }
        }
    }

    printf("\nShutting down...\n");
    return EXIT_SUCCESS;
}