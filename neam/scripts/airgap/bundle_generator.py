#!/usr/bin/env python3
"""
Offline Bundle Generator for NEAM Platform

This module provides functionality to generate offline deployment bundles
for air-gapped environments, including container images, Helm charts,
and platform configurations.
"""

import argparse
import hashlib
import json
import os
import shutil
import subprocess
import sys
import tarfile
import tempfile
import yaml
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple
from urllib.parse import urlparse
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class BundleType(Enum):
    """Types of artifacts that can be included in an offline bundle."""
    CONTAINER_IMAGE = "container_image"
    HELM_CHART = "helm_chart"
    CONFIG_BUNDLE = "config_bundle"
    KUSTOMIZE_BASE = "kustomize_base"
    OPERATOR_BUNDLE = "operator_bundle"


@dataclass
class ContainerImage:
    """Represents a container image to be exported."""
    repository: str
    tag: str
    digest: Optional[str] = None
    platform: str = "linux/amd64"
    registry: str = ""
    
    @property
    def full_name(self) -> str:
        """Get the fully qualified image name."""
        return f"{self.repository}:{self.tag}"
    
    @property
    def reference(self) -> str:
        """Get the image reference (with digest if available)."""
        if self.digest:
            return f"{repository}@{self.digest}"
        return self.full_name
    
    def to_dict(self) -> dict:
        return {
            "repository": self.repository,
            "tag": self.tag,
            "digest": self.digest,
            "platform": self.platform,
            "registry": self.registry
        }
    
    @classmethod
    def from_dict(cls, data: dict) -> 'ContainerImage':
        return cls(
            repository=data["repository"],
            tag=data["tag"],
            digest=data.get("digest"),
            platform=data.get("platform", "linux/amd64"),
            registry=data.get("registry", "")
        )


@dataclass
class HelmChart:
    """Represents a Helm chart to be exported."""
    name: str
    version: str
    repository: str
    chart_url: str
    values_file: Optional[str] = None
    dependencies: List[str] = field(default_factory=list)
    
    def to_dict(self) -> dict:
        return {
            "name": self.name,
            "version": self.version,
            "repository": self.repository,
            "chart_url": self.chart_url,
            "values_file": self.values_file,
            "dependencies": self.dependencies
        }
    
    @classmethod
    def from_dict(cls, data: dict) -> 'HelmChart':
        return cls(
            name=data["name"],
            version=data["version"],
            repository=data["repository"],
            chart_url=data["chart_url"],
            values_file=data.get("values_file"),
            dependencies=data.get("dependencies", [])
        )


@dataclass
class ConfigBundle:
    """Represents a configuration bundle to be exported."""
    name: str
    source_path: str
    config_type: str
    includes: List[str] = field(default_factory=list)
    transformers: List[str] = field(default_factory=list)
    
    def to_dict(self) -> dict:
        return {
            "name": self.name,
            "source_path": self.source_path,
            "config_type": self.config_type,
            "includes": self.includes,
            "transformers": self.transformers
        }
    
    @classmethod
    def from_dict(cls, data: dict) -> 'ConfigBundle':
        return cls(
            name=data["name"],
            source_path=data["source_path"],
            config_type=data["config_type"],
            includes=data.get("includes", []),
            transformers=data.get("transformers", [])
        )


@dataclass
class BundleManifest:
    """Manifest for an offline deployment bundle."""
    bundle_version: str
    bundle_type: str
    created_at: str
    platform_version: str
    target_environment: str
    container_images: List[dict] = field(default_factory=list)
    helm_charts: List[dict] = field(default_factory=list)
    config_bundles: List[dict] = field(default_factory=list)
    total_size: int = 0
    checksum: str = ""
    signatures: Dict[str, str] = field(default_factory=dict)
    metadata: Dict[str, str] = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        return {
            "bundle_version": self.bundle_version,
            "bundle_type": self.bundle_type,
            "created_at": self.created_at,
            "platform_version": self.platform_version,
            "target_environment": self.target_environment,
            "container_images": self.container_images,
            "helm_charts": self.helm_charts,
            "config_bundles": self.config_bundles,
            "total_size": self.total_size,
            "checksum": self.checksum,
            "signatures": self.signatures,
            "metadata": self.metadata
        }
    
    def save(self, path: str) -> None:
        """Save manifest to file."""
        with open(path, 'w') as f:
            yaml.dump(self.to_dict(), f, default_flow_style=False)
        logger.info(f"Manifest saved to {path}")
    
    @classmethod
    def load(cls, path: str) -> 'BundleManifest':
        """Load manifest from file."""
        with open(path, 'r') as f:
            data = yaml.safe_load(f)
        
        manifest = cls(
            bundle_version=data["bundle_version"],
            bundle_type=data["bundle_type"],
            created_at=data["created_at"],
            platform_version=data["platform_version"],
            target_environment=data["target_environment"]
        )
        manifest.container_images = data.get("container_images", [])
        manifest.helm_charts = data.get("helm_charts", [])
        manifest.config_bundles = data.get("config_bundles", [])
        manifest.total_size = data.get("total_size", 0)
        manifest.checksum = data.get("checksum", "")
        manifest.signatures = data.get("signatures", {})
        manifest.metadata = data.get("metadata", {})
        
        return manifest


class OfflineBundleGenerator:
    """
    Generator for offline deployment bundles.
    
    This class provides functionality to create complete offline deployment
    bundles containing container images, Helm charts, and platform configurations
    for air-gapped environment deployment.
    """
    
    def __init__(
        self,
        output_dir: str,
        platform_version: str = "1.0.0",
        target_environment: str = "air-gapped"
    ):
        self.output_dir = Path(output_dir)
        self.platform_version = platform_version
        self.target_environment = target_environment
        self.manifest = BundleManifest(
            bundle_version="1.0.0",
            bundle_type="complete",
            created_at=datetime.utcnow().isoformat(),
            platform_version=platform_version,
            target_environment=target_environment
        )
        self.temp_dir = None
        self.bundle_dir = None
        
        # Validate prerequisites
        self._check_prerequisites()
    
    def _check_prerequisites(self) -> None:
        """Check that required tools are available."""
        prerequisites = {
            "docker": "docker",
            "skopeo": "skopeo",
            "helm": "helm",
            "kubectl": "kubectl",
            "oras": "oras"
        }
        
        missing = []
        for tool, name in prerequisites.items():
            if not shutil.which(tool):
                missing.append(name)
        
        if missing:
            raise RuntimeError(
                f"Missing required tools: {', '.join(missing)}"
            )
        
        logger.info("All prerequisites met")
    
    def add_container_images(
        self,
        images: List[ContainerImage],
        skip_existing: bool = True
    ) -> int:
        """
        Add container images to the bundle.
        
        Args:
            images: List of ContainerImage objects to add
            skip_existing: Skip images that already exist in the bundle
            
        Returns:
            Number of images added
        """
        added_count = 0
        
        for image in images:
            if skip_existing and any(
                img["repository"] == image.repository and img["tag"] == image.tag
                for img in self.manifest.container_images
            ):
                logger.info(f"Skipping existing image: {image.full_name}")
                continue
            
            try:
                self._export_container_image(image)
                self.manifest.container_images.append(image.to_dict())
                added_count += 1
                logger.info(f"Added image: {image.full_name}")
            except Exception as e:
                logger.error(f"Failed to export image {image.full_name}: {e}")
                raise
        
        return added_count
    
    def _export_container_image(self, image: ContainerImage) -> str:
        """
        Export a container image to the bundle.
        
        Args:
            image: ContainerImage to export
            
        Returns:
            Path to the exported image
        """
        if not self.bundle_dir:
            raise RuntimeError("Bundle directory not initialized")
        
        # Create images directory
        images_dir = self.bundle_dir / "images"
        images_dir.mkdir(parents=True, exist_ok=True)
        
        # Determine output path
        image_dir = images_dir / image.repository.replace("/", "_")
        image_dir.mkdir(parents=True, exist_ok=True)
        
        # Export using skopeo
        output_path = image_dir / f"{image.tag}.tar"
        
        if output_path.exists() and output_path.stat().st_size > 0:
            logger.info(f"Image already exported: {image.full_name}")
            return str(output_path)
        
        # Source image reference
        if image.digest:
            source_ref = f"{image.repository}@{image.digest}"
        else:
            source_ref = f"{image.repository}:{image.tag}"
        
        # Use skopeo to copy image to tarball
        cmd = [
            "skopeo",
            "copy",
            f"docker://{source_ref}",
            f"docker-archive:{output_path}",
            "--dest-platform", image.platform
        ]
        
        logger.info(f"Exporting image: {source_ref} -> {output_path}")
        
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True
        )
        
        if result.returncode != 0:
            raise RuntimeError(
                f"Failed to export image {source_ref}: {result.stderr}"
            )
        
        # Get image digest if not provided
        if not image.digest:
            cmd = [
                "skopeo",
                "inspect",
                f"docker://{source_ref}",
                "--raw"
            ]
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True
            )
            if result.returncode == 0:
                import digestparser
                digest = digestparser.parse(result.stdout)
                if digest:
                    image.digest = str(digest)
        
        return str(output_path)
    
    def add_helm_charts(
        self,
        charts: List[HelmChart],
        skip_existing: bool = True
    ) -> int:
        """
        Add Helm charts to the bundle.
        
        Args:
            charts: List of HelmChart objects to add
            skip_existing: Skip charts that already exist in the bundle
            
        Returns:
            Number of charts added
        """
        added_count = 0
        
        for chart in charts:
            if skip_existing and any(
                ch["name"] == chart.name and ch["version"] == chart.version
                for ch in self.helm_charts
            ):
                logger.info(f"Skipping existing chart: {chart.name}-{chart.version}")
                continue
            
            try:
                self._export_helm_chart(chart)
                self.manifest.helm_charts.append(chart.to_dict())
                added_count += 1
                logger.info(f"Added chart: {chart.name}-{chart.version}")
            except Exception as e:
                logger.error(f"Failed to export chart {chart.name}: {e}")
                raise
        
        return added_count
    
    def _export_helm_chart(self, chart: HelmChart) -> str:
        """
        Export a Helm chart to the bundle.
        
        Args:
            chart: HelmChart to export
            
        Returns:
            Path to the exported chart
        """
        if not self.bundle_dir:
            raise RuntimeError("Bundle directory not initialized")
        
        # Create charts directory
        charts_dir = self.bundle_dir / "charts"
        charts_dir.mkdir(parents=True, exist_ok=True)
        
        # Determine output path
        output_path = charts_dir / f"{chart.name}-{chart.version}.tgz"
        
        if output_path.exists() and output_path.stat().st_size > 0:
            logger.info(f"Chart already exported: {chart.name}-{chart.version}")
            return str(output_path)
        
        # Use helm pull to download chart
        cmd = [
            "helm",
            "pull",
            chart.chart_url,
            "--destination", str(charts_dir),
            "--version", chart.version,
            "--repo", chart.repository
        ]
        
        logger.info(f"Downloading chart: {chart.chart_url}")
        
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True
        )
        
        if result.returncode != 0:
            raise RuntimeError(
                f"Failed to download chart {chart.name}: {result.stderr}"
            )
        
        # Rename if necessary
        downloaded_path = charts_dir / f"{chart.name}-{chart.version}.tgz"
        if downloaded_path != output_path:
            downloaded_path.rename(output_path)
        
        # Download dependencies if specified
        if chart.dependencies:
            self._download_chart_dependencies(chart, charts_dir)
        
        return str(output_path)
    
    def _download_chart_dependencies(
        self,
        chart: HelmChart,
        charts_dir: Path
    ) -> None:
        """Download chart dependencies."""
        # Create a temporary directory for dependency download
        temp_dir = tempfile.mkdtemp()
        
        try:
            # Download chart to temp dir
            cmd = [
                "helm",
                "pull",
                chart.chart_url,
                "--destination", temp_dir,
                "--version", chart.version
            ]
            
            subprocess.run(cmd, capture_output=True, check=True)
            
            # Extract and download dependencies
            chart_path = Path(temp_dir) / f"{chart.name}-{chart.version}.tgz"
            
            with tarfile.open(chart_path, 'r:gz') as tar:
                tar.extractall(temp_dir)
            
            # Download dependencies
            extracted_dir = Path(temp_dir) / chart.name
            
            # Create Chart.yaml with dependency repository
            chart_yaml_path = extracted_dir / "Chart.yaml"
            if chart_yaml_path.exists():
                # Pull dependencies
                subprocess.run(
                    ["helm", "dependency", "build", str(extracted_dir)],
                    capture_output=True,
                    check=True
                )
                
                # Move dependency charts to main charts directory
                charts_subdir = extracted_dir / "charts"
                if charts_subdir.exists():
                    for dep_chart in charts_subdir.glob("*.tgz"):
                        dep_chart.rename(charts_dir / dep_chart.name)
        
        finally:
            shutil.rmtree(temp_dir, ignore_errors=True)
    
    def add_config_bundle(
        self,
        config: ConfigBundle,
        base_path: str = "."
    ) -> None:
        """
        Add a configuration bundle to the bundle.
        
        Args:
            config: ConfigBundle to add
            base_path: Base path for resolving relative paths
        """
        if not self.bundle_dir:
            raise RuntimeError("Bundle directory not initialized")
        
        # Create configs directory
        configs_dir = self.bundle_dir / "configs"
        configs_dir.mkdir(parents=True, exist_ok=True)
        
        # Copy configuration files
        config_dest = configs_dir / config.name
        source_path = Path(base_path) / config.source_path
        
        if source_path.is_file():
            shutil.copy2(source_path, config_dest)
        elif source_path.is_dir():
            shutil.copytree(source_path, config_dest, dirs_exist_ok=True)
        else:
            raise FileNotFoundError(
                f"Configuration source path not found: {source_path}"
            )
        
        # Process transformers if specified
        for transformer in config.transformers:
            self._apply_transformer(config_dest, transformer)
        
        self.manifest.config_bundles.append(config.to_dict())
        logger.info(f"Added config bundle: {config.name}")
    
    def _apply_transformer(
        self,
        config_path: Path,
        transformer: str
    ) -> None:
        """
        Apply a configuration transformer.
        
        Args:
            config_path: Path to configuration
            transformer: Transformer name to apply
        """
        transformers = {
            "namespace-isolation": self._transform_namespace_isolation,
            "secret-masking": self._transform_secret_masking,
            "image-registry": self._transform_image_registry,
            "network-policy": self._transform_network_policy
        }
        
        if transformer in transformers:
            transformers[transformer](config_path)
        else:
            logger.warning(f"Unknown transformer: {transformer}")
    
    def _transform_namespace_isolation(self, config_path: Path) -> None:
        """Apply namespace isolation transformer."""
        # Replace external registry references with local
        for yaml_file in config_path.rglob("*.yaml"):
            try:
                with open(yaml_file, 'r') as f:
                    content = yaml.safe_load(f)
                
                if content and 'images' in content:
                    for img in content.get('images', []):
                        if 'registry' in img:
                            img['registry'] = 'local-registry.neam.internal:5000'
                
                with open(yaml_file, 'w') as f:
                    yaml.dump(content, f)
            except Exception as e:
                logger.warning(f"Failed to transform {yaml_file}: {e}")
    
    def _transform_secret_masking(self, config_path: Path) -> None:
        """Apply secret masking transformer."""
        # Placeholder for secret masking logic
        pass
    
    def _transform_image_registry(self, config_path: Path) -> None:
        """Apply image registry transformation."""
        for yaml_file in config_path.rglob("*.yaml"):
            try:
                with open(yaml_file, 'r') as f:
                    content = yaml.safe_load(f)
                
                if content:
                    # Replace image references
                    content_str = yaml.dump(content)
                    content_str = content_str.replace(
                        "gcr.io/",
                        "local-registry.neam.internal:5000/"
                    )
                    content_str = content_str.replace(
                        "docker.io/",
                        "local-registry.neam.internal:5000/"
                    )
                    content_str = content_str.replace(
                        "quay.io/",
                        "local-registry.neam.internal:5000/"
                    )
                    
                    content = yaml.safe_load(content_str)
                    
                    with open(yaml_file, 'w') as f:
                        yaml.dump(content, f)
            except Exception as e:
                logger.warning(f"Failed to transform {yaml_file}: {e}")
    
    def _transform_network_policy(self, config_path: Path) -> None:
        """Apply network policy transformation."""
        # Add default deny policies if not present
        pass
    
    def generate_manifest(self) -> BundleManifest:
        """Generate the bundle manifest."""
        if not self.bundle_dir:
            raise RuntimeError("Bundle directory not initialized")
        
        # Calculate total size
        total_size = 0
        for path in self.bundle_dir.rglob("*"):
            if path.is_file():
                total_size += path.stat().st_size
        
        self.manifest.total_size = total_size
        
        # Calculate checksum
        checksum = self._calculate_checksum()
        self.manifest.checksum = checksum
        
        # Save manifest
        manifest_path = self.bundle_dir / "bundle-manifest.yaml"
        self.manifest.save(str(manifest_path))
        
        return self.manifest
    
    def _calculate_checksum(self) -> str:
        """Calculate SHA256 checksum of bundle."""
        sha256_hash = hashlib.sha256()
        
        for path in sorted(self.bundle_dir.rglob("*")):
            if path.is_file():
                with open(path, "rb") as f:
                    for chunk in iter(lambda: f.read(8192), b""):
                        sha256_hash.update(chunk)
        
        return sha256_hash.hexdigest()
    
    def create_bundle(self) -> str:
        """
        Create the complete offline bundle.
        
        Returns:
            Path to the created bundle archive
        """
        # Initialize bundle directory
        timestamp = datetime.utcnow().strftime("%Y%m%d-%H%M%S")
        self.bundle_dir = self.output_dir / f"neam-bundle-{timestamp}"
        self.bundle_dir.mkdir(parents=True, exist_ok=True)
        
        logger.info(f"Creating bundle in: {self.bundle_dir}")
        
        # Generate manifest
        manifest = self.generate_manifest()
        
        # Create archive
        archive_path = self.output_dir / f"neam-bundle-{timestamp}.tar.gz"
        
        with tarfile.open(archive_path, "w:gz") as tar:
            tar.add(
                self.bundle_dir,
                arcname=os.path.basename(self.bundle_dir)
            )
        
        logger.info(f"Bundle created: {archive_path}")
        logger.info(f"Bundle size: {self._format_size(manifest.total_size)}")
        logger.info(f"Checksum: {manifest.checksum}")
        
        return str(archive_path)
    
    def _format_size(self, size: int) -> str:
        """Format size in human-readable format."""
        for unit in ['B', 'KB', 'MB', 'GB', 'TB']:
            if size < 1024:
                return f"{size:.2f} {unit}"
            size /= 1024
        return f"{size:.2f} PB"
    
    def cleanup(self) -> None:
        """Clean up temporary files."""
        if self.temp_dir and Path(self.temp_dir).exists():
            shutil.rmtree(self.temp_dir, ignore_errors=True)
            logger.info("Cleanup completed")


class BundleVerifier:
    """Verifier for offline deployment bundles."""
    
    def __init__(self, bundle_path: str):
        self.bundle_path = Path(bundle_path)
        self.manifest: Optional[BundleManifest] = None
        self.errors: List[str] = []
        self.warnings: List[str] = []
    
    def verify(self) -> bool:
        """
        Verify the bundle integrity.
        
        Returns:
            True if verification passed, False otherwise
        """
        self.errors = []
        self.warnings = []
        
        logger.info(f"Verifying bundle: {self.bundle_path}")
        
        # Check bundle exists
        if not self.bundle_path.exists():
            self.errors.append(f"Bundle not found: {self.bundle_path}")
            return False
        
        # Extract if archive
        if self.bundle_path.suffix == ".gz":
            extract_dir = self.bundle_path.parent / "verify-temp"
            self._extract_bundle(extract_dir)
            bundle_dir = extract_dir / self.bundle_path.stem.replace(".tar", "")
        else:
            bundle_dir = self.bundle_path
        
        # Load manifest
        manifest_path = bundle_dir / "bundle-manifest.yaml"
        if not manifest_path.exists():
            self.errors.append(f"Manifest not found: {manifest_path}")
            return False
        
        self.manifest = BundleManifest.load(str(manifest_path))
        
        # Run verification checks
        self._verify_checksum(bundle_dir)
        self._verify_container_images(bundle_dir)
        self._verify_helm_charts(bundle_dir)
        self._verify_config_bundles(bundle_dir)
        self._verify_signatures(bundle_dir)
        
        # Report results
        if self.errors:
            logger.error(f"Verification failed with {len(self.errors)} errors")
            for error in self.errors:
                logger.error(f"  ERROR: {error}")
        
        if self.warnings:
            logger.warning(f"Verification completed with {len(self.warnings)} warnings")
            for warning in self.warnings:
                logger.warning(f"  WARNING: {warning}")
        
        if not self.errors:
            logger.info("Bundle verification passed!")
            return True
        
        return False
    
    def _extract_bundle(self, extract_dir: Path) -> None:
        """Extract bundle archive."""
        if extract_dir.exists():
            shutil.rmtree(extract_dir)
        extract_dir.mkdir(parents=True)
        
        with tarfile.open(self.bundle_path, "r:gz") as tar:
            tar.extractall(extract_dir)
    
    def _verify_checksum(self, bundle_dir: Path) -> None:
        """Verify bundle checksum."""
        sha256_hash = hashlib.sha256()
        
        for path in sorted(bundle_dir.rglob("*")):
            if path.is_file() and path.name != "bundle-manifest.yaml":
                with open(path, "rb") as f:
                    for chunk in iter(lambda: f.read(8192), b""):
                        sha256_hash.update(chunk)
        
        calculated = sha256_hash.hexdigest()
        
        if calculated != self.manifest.checksum:
            self.errors.append(
                f"Checksum mismatch: expected {self.manifest.checksum}, "
                f"got {calculated}"
            )
        else:
            logger.info("Checksum verified")
    
    def _verify_container_images(self, bundle_dir: Path) -> None:
        """Verify container images."""
        images_dir = bundle_dir / "images"
        
        if not images_dir.exists():
            if self.manifest.container_images:
                self.errors.append("Images directory not found")
            return
        
        for image_data in self.manifest.container_images:
            image_path = images_dir / f"{image_data['repository'].replace('/', '_')}"
            tar_file = image_path / f"{image_data['tag']}.tar"
            
            if not tar_file.exists():
                self.errors.append(f"Image tar not found: {tar_file}")
                continue
            
            if tar_file.stat().st_size == 0:
                self.errors.append(f"Image tar is empty: {tar_file}")
                continue
            
            # Verify tarball is valid
            try:
                with tarfile.open(tar_file, 'r') as tar:
                    members = tar.getnames()
                    if not members:
                        self.warnings.append(f"Image tar may be empty: {tar_file}")
            except Exception as e:
                self.errors.append(f"Invalid image tar {tar_file}: {e}")
    
    def _verify_helm_charts(self, bundle_dir: Path) -> None:
        """Verify Helm charts."""
        charts_dir = bundle_dir / "charts"
        
        if not charts_dir.exists():
            if self.manifest.helm_charts:
                self.warnings.append("Charts directory not found")
            return
        
        for chart_data in self.manifest.helm_charts:
            chart_file = charts_dir / f"{chart_data['name']}-{chart_data['version']}.tgz"
            
            if not chart_file.exists():
                self.errors.append(f"Chart not found: {chart_file}")
                continue
            
            # Verify chart is valid
            try:
                with tarfile.open(chart_file, 'r:gz') as tar:
                    members = tar.getnames()
                    if 'Chart.yaml' not in members:
                        self.errors.append(f"Invalid chart (no Chart.yaml): {chart_file}")
            except Exception as e:
                self.errors.append(f"Invalid chart {chart_file}: {e}")
    
    def _verify_config_bundles(self, bundle_dir: Path) -> None:
        """Verify configuration bundles."""
        configs_dir = bundle_dir / "configs"
        
        if not configs_dir.exists():
            if self.manifest.config_bundles:
                self.warnings.append("Configs directory not found")
            return
        
        for config_data in self.manifest.config_bundles:
            config_path = configs_dir / config_data['name']
            
            if not config_path.exists():
                self.errors.append(f"Config bundle not found: {config_path}")
    
    def _verify_signatures(self, bundle_dir: Path) -> None:
        """Verify bundle signatures."""
        if not self.manifest.signatures:
            self.warnings.append("No signatures found in bundle")
            return
        
        # Verify each signature
        for signer, signature in self.manifest.signatures.items():
            # Signature verification would be implemented here
            # For now, just log that signatures exist
            logger.info(f"Signature found for: {signer}")


def main():
    """Main entry point for offline bundle generator."""
    parser = argparse.ArgumentParser(
        description="NEAM Platform Offline Bundle Generator"
    )
    
    subparsers = parser.add_subparsers(dest="command", help="Commands")
    
    # Generate command
    generate_parser = subparsers.add_parser("generate", help="Generate offline bundle")
    generate_parser.add_argument(
        "--output", "-o",
        default="./output",
        help="Output directory for bundle"
    )
    generate_parser.add_argument(
        "--images",
        help="JSON file containing list of container images"
    )
    generate_parser.add_argument(
        "--charts",
        help="JSON file containing list of Helm charts"
    )
    generate_parser.add_argument(
        "--configs",
        help="JSON file containing configuration bundles"
    )
    generate_parser.add_argument(
        "--platform-version",
        default="1.0.0",
        help="Platform version"
    )
    generate_parser.add_argument(
        "--environment",
        default="air-gapped",
        help="Target environment"
    )
    
    # Verify command
    verify_parser = subparsers.add_parser("verify", help="Verify bundle integrity")
    verify_parser.add_argument(
        "--bundle", "-b",
        required=True,
        help="Path to bundle archive or directory"
    )
    
    # List command
    list_parser = subparsers.add_parser("list", help="List bundle contents")
    list_parser.add_argument(
        "--bundle", "-b",
        required=True,
        help="Path to bundle archive or directory"
    )
    
    args = parser.parse_args()
    
    if args.command == "generate":
        # Create output directory
        output_dir = Path(args.output)
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Initialize generator
        generator = OfflineBundleGenerator(
            output_dir=str(output_dir),
            platform_version=args.platform_version,
            target_environment=args.environment
        )
        
        # Load and add container images
        if args.images:
            with open(args.images, 'r') as f:
                images_data = json.load(f)
            
            images = [ContainerImage.from_dict(img) for img in images_data]
            generator.add_container_images(images)
        
        # Load and add Helm charts
        if args.charts:
            with open(args.charts, 'r') as f:
                charts_data = json.load(f)
            
            charts = [HelmChart.from_dict(chart) for chart in charts_data]
            generator.add_helm_charts(charts)
        
        # Load and add config bundles
        if args.configs:
            with open(args.configs, 'r') as f:
                configs_data = json.load(f)
            
            for config_data in configs_data:
                config = ConfigBundle.from_dict(config_data)
                generator.add_config_bundle(config)
        
        # Create bundle
        bundle_path = generator.create_bundle()
        print(f"Bundle created: {bundle_path}")
        
        generator.cleanup()
    
    elif args.command == "verify":
        verifier = BundleVerifier(args.bundle)
        success = verifier.verify()
        sys.exit(0 if success else 1)
    
    elif args.command == "list":
        bundle_path = Path(args.bundle)
        
        if bundle_path.suffix == ".gz":
            import tempfile
            temp_dir = tempfile.mkdtemp()
            with tarfile.open(bundle_path, "r:gz") as tar:
                tar.extractall(temp_dir)
            bundle_path = temp_dir / bundle_path.stem.replace(".tar", "")
        
        manifest_path = bundle_path / "bundle-manifest.yaml"
        if manifest_path.exists():
            manifest = BundleManifest.load(str(manifest_path))
            print(f"Bundle: {manifest.bundle_type}")
            print(f"Version: {manifest.bundle_version}")
            print(f"Created: {manifest.created_at}")
            print(f"Platform: {manifest.platform_version}")
            print(f"Environment: {manifest.target_environment}")
            print(f"Size: {manifest.total_size} bytes")
            print(f"Checksum: {manifest.checksum}")
            print()
            print(f"Container Images: {len(manifest.container_images)}")
            for img in manifest.container_images:
                print(f"  - {img['repository']}:{img['tag']}")
            print()
            print(f"Helm Charts: {len(manifest.helm_charts)}")
            for chart in manifest.helm_charts:
                print(f"  - {chart['name']}-{chart['version']}")
            print()
            print(f"Config Bundles: {len(manifest.config_bundles)}")
            for config in manifest.config_bundles:
                print(f"  - {config['name']}")
        else:
            print("Manifest not found")
            sys.exit(1)
    
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
