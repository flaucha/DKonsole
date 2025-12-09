export const getResourceTemplate = (kind, namespace) => {
    const ns = namespace || 'dkonsole';

    switch (kind) {
        case 'Namespace':
            return `apiVersion: v1
kind: Namespace
metadata:
  name: example-namespace
  labels:
    app: example
`;
        case 'Deployment':
            return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
  namespace: ${ns}
  labels:
    app: example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
        - name: example-container
          image: nginx:latest
          ports:
            - containerPort: 80
          resources:
            requests:
              memory: "64Mi"
              cpu: "100m"
            limits:
              memory: "128Mi"
              cpu: "200m"
`;
        case 'Service':
            return `apiVersion: v1
kind: Service
metadata:
  name: example-service
  namespace: ${ns}
  labels:
    app: example
spec:
  type: ClusterIP
  selector:
    app: example
  ports:
    - name: http
      port: 80
      targetPort: 80
      protocol: TCP
`;
        case 'Pod':
            return `apiVersion: v1
kind: Pod
metadata:
  name: example-pod
  namespace: ${ns}
  labels:
    app: example
spec:
  containers:
    - name: example-container
      image: nginx:latest
      ports:
        - containerPort: 80
      resources:
        requests:
          memory: "64Mi"
          cpu: "100m"
        limits:
          memory: "128Mi"
          cpu: "200m"
  restartPolicy: Always
`;
        case 'ConfigMap':
            return `apiVersion: v1
kind: ConfigMap
metadata:
  name: example-configmap
  namespace: ${ns}
data:
  # Add your configuration data here
  config.properties: |
    key1=value1
    key2=value2
  another.conf: |
    setting=enabled
`;
        case 'Secret':
            return `apiVersion: v1
kind: Secret
metadata:
  name: example-secret
  namespace: ${ns}
type: Opaque
stringData:
  # Add your secret data here (will be base64 encoded automatically)
  username: admin
  password: changeme
`;
        case 'Ingress':
            return `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
  namespace: ${ns}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
    - host: example.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: example-service
                port:
                  number: 80
`;
        case 'Job':
            return `apiVersion: batch/v1
kind: Job
metadata:
  name: example-job
  namespace: ${ns}
spec:
  completions: 1
  parallelism: 1
  backoffLimit: 3
  template:
    spec:
      containers:
        - name: example-job
          image: busybox:latest
          command: ["echo", "Hello from Job"]
      restartPolicy: Never
`;
        case 'CronJob':
            return `apiVersion: batch/v1
kind: CronJob
metadata:
  name: example-cronjob
  namespace: ${ns}
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: example-cronjob
              image: busybox:latest
              command: ["echo", "Hello from CronJob"]
          restartPolicy: OnFailure
`;
        case 'StatefulSet':
            return `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: example-statefulset
  namespace: ${ns}
spec:
  serviceName: example-headless
  replicas: 1
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
        - name: example-container
          image: nginx:latest
          ports:
            - containerPort: 80
`;
        case 'DaemonSet':
            return `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: example-daemonset
  namespace: ${ns}
spec:
  selector:
    matchLabels:
      app: example
  template:
    metadata:
      labels:
        app: example
    spec:
      containers:
        - name: example-container
          image: nginx:latest
`;
        case 'PersistentVolumeClaim':
            return `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: example-pvc
  namespace: ${ns}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  # storageClassName: standard
`;
        case 'HorizontalPodAutoscaler':
            return `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: example-hpa
  namespace: ${ns}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: example-deployment
  minReplicas: 1
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 80
`;
        default:
            return `apiVersion: v1
kind: ${kind}
metadata:
  name: example-${kind.toLowerCase()}
${namespace ? `  namespace: ${namespace}` : ''}
`;
    }
};
