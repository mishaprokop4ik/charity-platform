description = "Gralde for Kurajj charity platform"
version = "0.0.1"

tasks.register("serverBuild") {
    group = "build"
    description = "Builds binary of project"
    val binaryDir: String by extra { properties.getOrDefault("dir", "bin") as String }
    val binaryName: String by extra { properties.getOrDefault("name", "kurajj_server") as String }
    doLast {
        exec {
            commandLine = listOf("go", "build", "-o", "./${binaryDir}/${binaryName}", "-a", ".")
        }
    }
}

tasks.register("serverDockerBuild") {
    group = "docker"
    description = "Builds server image by Dockerfile"
    val version: Int by extra { properties.getOrDefault("v", 1) as Int }
    val kvServerDockerName: String by extra { properties.getOrDefault("dockerName", "miprokop/kurajj_charity_planform") as String }
    doLast {
        exec {
            commandLine = listOf("docker", "build", "-t", "${kvServerDockerName}:v$version", "-f", "server/Dockerfile", ".")
        }
    }
}

tasks.register("dbDockerStart") {
    group = "docker"
    description = "Start PostgreSQL database in the Docker contrainer"
    val currentDir = System.getProperty("user.dir")
    doLast {
        exec {
            commandLine = listOf("docker", "run", "-v", "${currentDir}/db_data/:/var/lib/postgresql/data", "--name=kurajj_db", "-e", "POSTGRES_PASSWORD=rootroot", "-e", "POSTGRES_DB=kurajj", "-p", "5433:5432", "-d", "--rm", "postgres")
        }
    }
}

tasks.register("serverDockerPush") {
    group = "docker"
    description = "Pushes the Kurajj Charity Planform docker image to Dockerhub"
    val version: Int by extra { properties.getOrDefault("v", 1) as Int }
    val kvServerDockerName: String by extra { properties.getOrDefault("dockerName", "miprokop/kurajj_charity_planform") as String }
    doLast {
        exec {
            commandLine = listOf("docker", "push",  "${kvServerDockerName}:v$version")
        }
    }
}

tasks.register("addMigration") {
    group = "migration"
    description = "Create new SQL migration"

    val fileLength: String by extra { properties.getOrDefault("length", "14") as String }
    val name: String by extra { properties.get("migrationName") as String }
    val savePath: String by extra { properties.getOrDefault("savePath", "internal/repository/postgres/migrations") as String }
    doLast {
        exec {
            commandLine = listOf("migrate", "create", "-ext", "sql", "-dir", "${savePath}", "-seq", "-digits", "${fileLength}", "${name}")
        }
    }
}

tasks.register("migrateUp") {
    group = "migration"
    description = "Up migrations in spec dir"

    val migrationFilesPath: String by extra { properties.getOrDefault("migrationFilesPath", "internal/repository/postgres/migrations") as String }
    val databaseURL: String by extra { properties.getOrDefault("url", "postgres://postgres:rootroot@localhost:5433/kurajj?sslmode=disable") as String }
    doLast {
        exec {
            commandLine = listOf("migrate", "-path", "$migrationFilesPath", "-database", "$databaseURL", "up")
        }
    }
}
