You are tasked with implementing a backend service for Widgets Inc.

- Widgets have the following properties: Name, Serial number, List of connection ports
- Widgets rely on other widgets to work by associating their connection ports
- Each widget can have 0 to 3 connections to other widgets through its connection ports
- Connection ports have 3 types: P, R, and Q. Not all widgets have all port types
- A connection port can only associate with one Widget at a time

The API should support the following functionalities:

- Adding and removing widgets
- Associating specific widgets with other specific widgets using their ports
- The technology stack is not specified, but it should be scalable to handle approximately 10,000,000 widgets
- On average, each widget connects to 2 other widgets
- The system should handle frequent widget updates, potentially thousands of times per second

You do not need to provide a working solution. A rough concept code that can be the foundation is enough, but do try to cover all the implementation levels from database schema to deployment.
